package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/repositories"
	"time"
)

// WhisperRepositoryInterface defines the methods required from the Whisper repository
type WhisperRepositoryInterface interface {
	Transcribe(wavPath string, modelPath string, lang string) (string, error)
	TranscribeFull(wavPath string, modelPath string, lang string) (string, error)
}

type TranscribeUseCase interface {
	Execute(inputPath string, fileName string) (string, error)
}

type transcribeUseCase struct {
	whisperRepo         WhisperRepositoryInterface
	whisperProxyUsecase *WhisperProxyUsecase
	ragUsecase          *ragUsecases.RAGUsecase
	taskRepo            repositories.TranscriptionTaskRepository
	config              *utils.Config
}

func NewTranscribeUseCase(
	whisperRepo WhisperRepositoryInterface,
	whisperProxyUsecase *WhisperProxyUsecase,
	ragUsecase *ragUsecases.RAGUsecase,
	taskRepo repositories.TranscriptionTaskRepository,
	config *utils.Config,
) TranscribeUseCase {
	return &transcribeUseCase{
		whisperRepo:         whisperRepo,
		whisperProxyUsecase: whisperProxyUsecase,
		ragUsecase:          ragUsecase,
		taskRepo:            taskRepo,
		config:              config,
	}
}

func (uc *transcribeUseCase) Execute(inputPath string, fileName string) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcribe: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	taskID := utils.GenerateUUID()
	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if err := uc.taskRepo.SaveShortTask(taskID, status); err != nil {
		return "", err
	}

	utils.LogInfo("Transcribe: Started task %s for file %s", taskID, fileName)

	go uc.processAsync(taskID, inputPath)

	return taskID, nil
}

func (uc *transcribeUseCase) processAsync(taskID string, inputPath string) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("Transcribe Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", nil)
		}
	}()

	var text string
	var lang string
	var err error
	var usedPath string

	// Try PPU first
	if uc.whisperProxyUsecase != nil {
		if proxyErr := uc.whisperProxyUsecase.HealthCheck(); proxyErr == nil {
			result, fetchErr := uc.whisperProxyUsecase.FetchToOutsystems(inputPath, filepath.Base(inputPath))
			if fetchErr == nil && result != nil {
				text = result.Transcription
				lang = "id"
				usedPath = "PPU (Outsystems)"
			}
		}
	}

	// Fallback to local
	if text == "" {
		text, lang, err = uc.transcribeLocal(inputPath)
		usedPath = "Local Whisper (whisper.cpp)"
		if err != nil {
			utils.LogError("Transcribe Task %s: Local transcription failed: %v", taskID, err)
			uc.updateStatus(taskID, "failed", nil)
			return
		}
	}

	utils.LogInfo("Transcribe Task %s: Finished using %s", taskID, usedPath)

	// Refine (Grammar/Spelling)
	refined, _ := uc.ragUsecase.Refine(text, lang)

	result := &speechdtos.AsyncTranscriptionResultDTO{
		Transcription:    text,
		RefinedText:      refined,
		DetectedLanguage: lang,
	}

	uc.updateStatus(taskID, "completed", result)
}

func (uc *transcribeUseCase) transcribeLocal(inputPath string) (string, string, error) {
	utils.LogDebug("Speech: Starting local transcription via whisper.cpp...")
	tempDir := filepath.Dir(inputPath)

	wavPath := filepath.Join(tempDir, "processed.wav")
	if err := utils.ConvertToWav(inputPath, wavPath); err != nil {
		return "", "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	modelPath := uc.config.WhisperModelPath

	text, err := uc.whisperRepo.TranscribeFull(wavPath, modelPath, "id")
	if err != nil {
		return "", "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, "id", nil
}

func (uc *transcribeUseCase) updateStatus(taskID string, statusStr string, result *speechdtos.AsyncTranscriptionResultDTO) {
	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    statusStr,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}
	_ = uc.taskRepo.SaveShortTask(taskID, status)
}
