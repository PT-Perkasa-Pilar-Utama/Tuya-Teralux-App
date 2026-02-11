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

type TranscribeWhisperCppUseCase interface {
	Execute(inputPath string, fileName string, lang string) (string, error)
}

type transcribeWhisperCppUseCase struct {
	whisperRepo WhisperRepositoryInterface
	ragUsecase  *ragUsecases.RAGUsecase
	taskRepo    repositories.TranscriptionTaskRepository
	config      *utils.Config
}

func NewTranscribeWhisperCppUseCase(
	whisperRepo WhisperRepositoryInterface,
	ragUsecase *ragUsecases.RAGUsecase,
	taskRepo repositories.TranscriptionTaskRepository,
	config *utils.Config,
) TranscribeWhisperCppUseCase {
	return &transcribeWhisperCppUseCase{
		whisperRepo: whisperRepo,
		ragUsecase:  ragUsecase,
		taskRepo:    taskRepo,
		config:      config,
	}
}

func (uc *transcribeWhisperCppUseCase) Execute(inputPath string, fileName string, lang string) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("TranscribeWhisperCpp: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	taskID := utils.GenerateUUID()
	status := &speechdtos.AsyncTranscriptionLongStatusDTO{
		Status:    "pending",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if err := uc.taskRepo.SaveLongTask(taskID, status); err != nil {
		return "", err
	}

	utils.LogInfo("TranscribeWhisperCpp: Started task %s for file %s", taskID, fileName)

	go uc.processAsync(taskID, inputPath, lang)

	return taskID, nil
}

func (uc *transcribeWhisperCppUseCase) processAsync(taskID string, inputPath string, lang string) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("TranscribeWhisperCpp Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", nil)
		}
	}()

	text, err := uc.transcribeLongLocal(inputPath, lang)
	if err != nil {
		utils.LogError("TranscribeWhisperCpp Task %s: Transcription failed: %v", taskID, err)
		uc.updateStatus(taskID, "failed", nil)
		return
	}

	utils.LogInfo("TranscribeWhisperCpp Task %s: Transcription finished", taskID)

	// Refine
	refined, _ := uc.ragUsecase.Refine(text, lang)

	result := &speechdtos.AsyncTranscriptionLongResultDTO{
		Transcription:    text,
		RefinedText:      refined,
		DetectedLanguage: lang,
	}

	uc.updateStatus(taskID, "completed", result)
}

func (uc *transcribeWhisperCppUseCase) transcribeLongLocal(inputPath string, lang string) (string, error) {
	utils.LogDebug("Speech: Starting local LONG transcription via whisper.cpp...")
	tempDir := filepath.Dir(inputPath)

	wavPath := filepath.Join(tempDir, "processed_long.wav")
	if err := utils.ConvertToWav(inputPath, wavPath); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	modelPath := uc.config.WhisperModelPath

	text, err := uc.whisperRepo.TranscribeFull(wavPath, modelPath, lang)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, nil
}

func (uc *transcribeWhisperCppUseCase) updateStatus(taskID string, statusStr string, result *speechdtos.AsyncTranscriptionLongResultDTO) {
	status := &speechdtos.AsyncTranscriptionLongStatusDTO{
		Status:    statusStr,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}
	_ = uc.taskRepo.SaveLongTask(taskID, status)
}
