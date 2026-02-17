package usecases

import (
	"encoding/json"
	"fmt"
	"os"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/utilities"
	"time"
)

type TranscriptionMetadata struct {
	UID       string
	TeraluxID string
	Source    string // "mqtt", "rest", etc.
}

type TranscribeUseCase interface {
	TranscribeAudio(inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error)
}

type transcribeUseCase struct {
	whisperClient utilities.WhisperClient
	refineUC      ragUsecases.RefineUseCase
	cache         *tasks.BadgerTaskCache
	config        *utils.Config
	mqttSvc       *infrastructure.MqttService
}

func NewTranscribeUseCase(
	whisperClient utilities.WhisperClient,
	refineUC ragUsecases.RefineUseCase,
	cache *tasks.BadgerTaskCache,
	config *utils.Config,
	mqttSvc *infrastructure.MqttService,
) TranscribeUseCase {
	return &transcribeUseCase{
		whisperClient: whisperClient,
		refineUC:      refineUC,
		cache:         cache,
		config:        config,
		mqttSvc:       mqttSvc,
	}
}

func (uc *transcribeUseCase) TranscribeAudio(inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error) {
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

	var meta *TranscriptionMetadata
	if len(metadata) > 0 {
		meta = &metadata[0]
	}

	go uc.processAsync(taskID, inputPath, language, meta)

	return taskID, nil
}

func (uc *transcribeUseCase) processAsync(taskID string, inputPath string, reqLanguage string, metadata *TranscriptionMetadata) {
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
	// Priority: Use requested language if explicitly provided (e.g. from App), otherwise fallback to detected.
	refineLang := result.DetectedLanguage
	if reqLanguage != "" {
		refineLang = reqLanguage
	}
	refined, _ := uc.refineUC.RefineText(result.Transcription, refineLang)

	finalResult := &speechdtos.AsyncTranscriptionResultDTO{
		Transcription:    result.Transcription,
		RefinedText:      refined,
		DetectedLanguage: result.DetectedLanguage, // Keep original detection for record
	}

	uc.updateStatus(taskID, "completed", finalResult)

	// Chaining to /chat ONLY if initiated via MQTT
	if metadata != nil && metadata.Source == "mqtt" && metadata.TeraluxID != "" && uc.mqttSvc != nil {
		chatTopic := "users/teralux/chat"
		prompt := finalResult.RefinedText
		if prompt == "" {
			prompt = finalResult.Transcription
		}

		chatReq := map[string]string{
			"prompt":     prompt,
			"teralux_id": metadata.TeraluxID,
			"language":   result.DetectedLanguage,
			"uid":        metadata.UID,
		}
		payload, _ := json.Marshal(chatReq)
		uc.mqttSvc.Publish(chatTopic, 0, false, payload)
		utils.LogInfo("Transcribe Task %s: Chained result to %s", taskID, chatTopic)
	}
}

func (uc *transcribeUseCase) updateStatus(taskID string, statusStr string, result *speechdtos.AsyncTranscriptionResultDTO) {
	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    statusStr,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
