package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
	"time"
)

// WhisperCppRepositoryInterface defines the methods for local transcription
type WhisperCppRepositoryInterface interface {
	Transcribe(wavPath string, modelPath string, lang string) (string, error)
	TranscribeFull(wavPath string, modelPath string, lang string) (string, error)
}

// WhisperOrionRepositoryInterface defines the methods for remote Orion transcription
type WhisperOrionRepositoryInterface interface {
	Transcribe(audioPath string, lang string) (string, error)
}

type TranscribeUseCase interface {
	TranscribeAudio(inputPath string, fileName string, language string) (string, error)
}

type transcribeUseCase struct {
	whisperCpp          WhisperCppRepositoryInterface
	whisperOrion        WhisperOrionRepositoryInterface
	whisperProxyUsecase WhisperProxyUsecase
	refineUC            ragUsecases.RefineUseCase
	cache               *tasks.BadgerTaskCache
	config              *utils.Config
}

func NewTranscribeUseCase(
	whisperCpp WhisperCppRepositoryInterface,
	whisperOrion WhisperOrionRepositoryInterface,
	whisperProxyUsecase WhisperProxyUsecase,
	refineUC ragUsecases.RefineUseCase,
	cache *tasks.BadgerTaskCache,
	config *utils.Config,
) TranscribeUseCase {
	return &transcribeUseCase{
		whisperCpp:          whisperCpp,
		whisperOrion:        whisperOrion,
		whisperProxyUsecase: whisperProxyUsecase,
		refineUC:            refineUC,
		cache:               cache,
		config:              config,
	}
}

func (uc *transcribeUseCase) TranscribeAudio(inputPath string, fileName string, language string) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcribe: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	taskID := utils.GenerateUUID()
	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	// Mark as pending
	_ = uc.cache.Set(taskID, status)

	utils.LogInfo("Transcribe: Started task %s for file %s", taskID, fileName)

	go uc.processAsync(taskID, inputPath, language)

	return taskID, nil
}

func (uc *transcribeUseCase) processAsync(taskID string, inputPath string, reqLanguage string) {
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
			result, fetchErr := uc.whisperProxyUsecase.FetchToOutsystems(inputPath, filepath.Base(inputPath), reqLanguage)
			if fetchErr == nil && result != nil {
				text = result.Transcription
				lang = result.DetectedLanguage
				if lang == "" {
					lang = reqLanguage
				}
				usedPath = "PPU (Outsystems)"
			}
		}
	}

	// Try Orion Whisper
	if text == "" && uc.whisperOrion != nil && uc.config.WhisperServerURL != "" {
		res, err := uc.whisperOrion.Transcribe(inputPath, reqLanguage)
		if err == nil {
			text = res
			lang = reqLanguage
			usedPath = "Orion Whisper"
		}
	}

	// Fallback to local Cpp
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
	refined, _ := uc.refineUC.RefineText(text, lang)

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

	text, err := uc.whisperCpp.TranscribeFull(wavPath, modelPath, "id")
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
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
