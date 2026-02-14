package usecases

import (
	"fmt"
	"os"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/utilities"
	"time"
)

type TranscribeUseCase interface {
	TranscribeAudio(inputPath string, fileName string, language string) (string, error)
}

type transcribeUseCase struct {
	whisperClient utilities.WhisperClient
	refineUC      ragUsecases.RefineUseCase
	cache         *tasks.BadgerTaskCache
	config        *utils.Config
}

func NewTranscribeUseCase(
	whisperClient utilities.WhisperClient,
	refineUC ragUsecases.RefineUseCase,
	cache *tasks.BadgerTaskCache,
	config *utils.Config,
) TranscribeUseCase {
	return &transcribeUseCase{
		whisperClient: whisperClient,
		refineUC:      refineUC,
		cache:         cache,
		config:        config,
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

	// Use unified whisper client with automatic fallback
	result, err := uc.whisperClient.Transcribe(inputPath, reqLanguage)
	if err != nil {
		utils.LogError("Transcribe Task %s: All transcription methods failed: %v", taskID, err)
		uc.updateStatus(taskID, "failed", nil)
		return
	}

	utils.LogInfo("Transcribe Task %s: Finished using %s", taskID, result.Source)

	// Refine (Grammar/Spelling)
	refined, _ := uc.refineUC.RefineText(result.Transcription, result.DetectedLanguage)

	finalResult := &speechdtos.AsyncTranscriptionResultDTO{
		Transcription:    result.Transcription,
		RefinedText:      refined,
		DetectedLanguage: result.DetectedLanguage,
	}

	uc.updateStatus(taskID, "completed", finalResult)
}

func (uc *transcribeUseCase) updateStatus(taskID string, statusStr string, result *speechdtos.AsyncTranscriptionResultDTO) {
	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    statusStr,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
