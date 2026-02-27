package usecases

import (
	"encoding/json"
	"fmt"
	"os"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	ragUsecases "sensio/domain/rag/usecases"
	speechdtos "sensio/domain/speech/dtos"
	"time"
)

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// WhisperClient is the unified interface for all whisper transcription services
type WhisperClient interface {
	Transcribe(audioPath string, language string, diarize bool) (*speechdtos.WhisperResult, error)
}

type TranscriptionMetadata struct {
	UID         string
	TerminalID  string
	Source      string // "mqtt", "rest", etc.
	Trigger     string // e.g., "/api/speech/transcribe"
	DeleteAfter bool   // Whether to delete the audio file after processing
	Diarize     bool   // Whether to perform speaker diarization
}

type TranscribeUseCase interface {
	TranscribeAudio(inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error)
}

type transcribeUseCase struct {
	whisperClient  WhisperClient
	fallbackClient WhisperClient
	refineUC       ragUsecases.RefineUseCase
	store          *tasks.StatusStore[speechdtos.AsyncTranscriptionStatusDTO]
	cache          *tasks.BadgerTaskCache
	config         *utils.Config
	mqttSvc        mqttPublisher
}

func NewTranscribeUseCase(
	whisperClient WhisperClient,
	fallbackClient WhisperClient,
	refineUC ragUsecases.RefineUseCase,
	store *tasks.StatusStore[speechdtos.AsyncTranscriptionStatusDTO],
	cache *tasks.BadgerTaskCache,
	config *utils.Config,
	mqttSvc mqttPublisher,
) TranscribeUseCase {
	return &transcribeUseCase{
		whisperClient:  whisperClient,
		fallbackClient: fallbackClient,
		refineUC:       refineUC,
		store:          store,
		cache:          cache,
		config:         config,
		mqttSvc:        mqttSvc,
	}
}

func (uc *transcribeUseCase) TranscribeAudio(inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcribe: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	taskID := utils.GenerateUUID()
	var meta *TranscriptionMetadata
	if len(metadata) > 0 {
		meta = &metadata[0]
	}

	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		Trigger:   "",
		StartedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if meta != nil && meta.Trigger != "" {
		status.Trigger = meta.Trigger
	}

	// Mark as pending
	uc.store.Set(taskID, status)
	_ = uc.cache.Set(taskID, status)

	utils.LogInfo("Transcribe: Started task %s for file %s", taskID, fileName)

	go uc.processAsync(taskID, inputPath, language, meta)

	return taskID, nil
}

func (uc *transcribeUseCase) processAsync(taskID string, inputPath string, reqLanguage string, metadata *TranscriptionMetadata) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("Transcribe Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", nil, fmt.Errorf("internal panic: %v", r))
		}
	}()

	// Use unified whisper client with automatic fallback
	diarize := false
	if metadata != nil {
		diarize = metadata.Diarize
	}

	result, err := uc.whisperClient.Transcribe(inputPath, reqLanguage, diarize)
	if err != nil && uc.fallbackClient != nil {
		utils.LogWarn("Transcribe: Primary client failed, falling back to local: %v", err)
		result, err = uc.fallbackClient.Transcribe(inputPath, reqLanguage, diarize)
	}

	if err != nil {
		utils.LogError("Transcribe Task %s: All transcription methods failed: %v", taskID, err)
		uc.updateStatus(taskID, "failed", nil, err)

		// Clean up on failure if needed
		if metadata != nil && metadata.DeleteAfter {
			_ = os.Remove(inputPath)
			utils.LogDebug("Transcribe Task %s: Deleted temporary file %s after failure", taskID, inputPath)
		}
		return
	}

	utils.LogInfo("Transcribe Task %s: Finished using %s", taskID, result.Source)

	// Clean up permanent file if requested
	if metadata != nil && metadata.DeleteAfter {
		_ = os.Remove(inputPath)
		utils.LogDebug("Transcribe Task %s: Deleted temporary file %s after completion", taskID, inputPath)
	}

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

	uc.updateStatus(taskID, "completed", finalResult, nil)

	// Chaining to /chat ONLY if initiated via MQTT
	if metadata != nil && metadata.Source == "mqtt" && metadata.TerminalID != "" && uc.mqttSvc != nil {
		chatTopic := fmt.Sprintf("users/%s/chat", metadata.TerminalID)
		prompt := finalResult.RefinedText
		if prompt == "" {
			prompt = finalResult.Transcription
		}

		chatReq := map[string]string{
			"prompt":      prompt,
			"terminal_id": metadata.TerminalID,
			"language":    result.DetectedLanguage,
			"uid":         metadata.UID,
		}
		payload, _ := json.Marshal(chatReq)
		if err := uc.mqttSvc.Publish(chatTopic, 0, false, payload); err != nil {
			utils.LogError("TranscribeUseCase: Failed to publish transcript to MQTT: %v", err)
		}
		utils.LogInfo("Transcribe Task %s: Chained result to %s", taskID, chatTopic)
	}
}

func (uc *transcribeUseCase) updateStatus(taskID string, statusStr string, result *speechdtos.AsyncTranscriptionResultDTO, err error) {
	// Try to get existing status to preserve StartedAt
	var existing speechdtos.AsyncTranscriptionStatusDTO
	_, _, _ = uc.cache.GetWithTTL(taskID, &existing)

	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    statusStr,
		Result:    result,
		StartedAt: existing.StartedAt,
		Trigger:   existing.Trigger,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if err != nil {
		status.Error = err.Error()
		status.HTTPStatusCode = utils.GetErrorStatusCode(err)
	} else if statusStr == "completed" {
		status.HTTPStatusCode = 200
	}

	// Calculate duration if finished
	if statusStr == "completed" || statusStr == "failed" {
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			status.DurationSeconds = time.Since(startTime).Seconds()
		}
	}

	uc.store.Set(taskID, status)
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
