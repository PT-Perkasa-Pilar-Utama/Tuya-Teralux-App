package usecases

import (
	"fmt"
	"os"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
	"time"
)

type TranscribeWhisperCppUseCase interface {
	TranscribeWhisperCpp(inputPath string, fileName string, lang string) (string, error)
}

type transcribeWhisperCppUseCase struct {
	whisperClient WhisperClient
	refineUC      ragUsecases.RefineUseCase
	cache         *tasks.BadgerTaskCache
	config        *utils.Config
}

func NewTranscribeWhisperCppUseCase(
	whisperClient WhisperClient,
	refineUC ragUsecases.RefineUseCase,
	cache *tasks.BadgerTaskCache,
	config *utils.Config,
) TranscribeWhisperCppUseCase {
	return &transcribeWhisperCppUseCase{
		whisperClient: whisperClient,
		refineUC:      refineUC,
		cache:         cache,
		config:        config,
	}
}

func (uc *transcribeWhisperCppUseCase) TranscribeWhisperCpp(inputPath string, fileName string, lang string) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("TranscribeWhisperCpp: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	taskID := utils.GenerateUUID()
	status := &speechdtos.AsyncTranscriptionLongStatusDTO{
		Status:    "pending",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	_ = uc.cache.Set(taskID, status)

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

	// Transcribe with local whisper.cpp (via service)
	// Service handles conversion to WAV if needed, but we already converted it here?
	// WhisperLocalService checks extension and converts if needed.
	// But here strict WAV is expected by old repo logic?
	// Service accepts audioPath.
	// Let's pass the inputPath directly to service and let it handle conversion!
	// Removing local conversion block?
	// Wait, processAsync converted to 'processed.wav'.
	// If I pass inputPath to service, service converts to 'processed_local_service.wav'.
	// It's fine. Redundant conversion but safe.
	// Actually, better to just call service with inputPath.
	
	result, err := uc.whisperClient.Transcribe(inputPath, lang)
	if err != nil {
		utils.LogError("TranscribeWhisperCpp Task %s: Transcription failed: %v", taskID, err)
		uc.updateStatus(taskID, "failed", nil)
		return
	}
	text := result.Transcription

	utils.LogInfo("TranscribeWhisperCpp Task %s: Transcription finished", taskID)

	// Refine
	refined, _ := uc.refineUC.RefineText(text, lang)

	finalResult := &speechdtos.AsyncTranscriptionLongResultDTO{
		Transcription:    text,
		RefinedText:      refined,
		DetectedLanguage: lang,
	}

	uc.updateStatus(taskID, "completed", finalResult)
}

func (uc *transcribeWhisperCppUseCase) updateStatus(taskID string, statusStr string, result *speechdtos.AsyncTranscriptionLongResultDTO) {
	status := &speechdtos.AsyncTranscriptionLongStatusDTO{
		Status:    statusStr,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
