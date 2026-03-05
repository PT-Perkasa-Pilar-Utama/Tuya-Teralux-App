package usecases

import (
	"context"
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
	Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*speechdtos.WhisperResult, error)
}

type TranscriptionMetadata struct {
	UID            string
	TerminalID     string
	Source         string // "mqtt", "rest", etc.
	Trigger        string // e.g., "/api/speech/transcribe"
	DeleteAfter    bool   // Whether to delete the audio file after processing
	Diarize        bool   // Whether to perform speaker diarization
	IdempotencyKey string // Client-provided idempotency key
}

type TranscribeUseCase interface {
	TranscribeAudio(ctx context.Context, inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error)
	TranscribeAudioSync(ctx context.Context, inputPath string, language string, diarize bool, refine bool) (*speechdtos.AsyncTranscriptionResultDTO, error)
	CheckIdempotency(idempotencyKey string, audioHash string, language string, terminalID string) (string, bool)
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

func (uc *transcribeUseCase) CheckIdempotency(idempotencyKey string, audioHash string, language string, terminalID string) (string, bool) {
	if idempotencyKey == "" {
		return "", false
	}
	hashInput := fmt.Sprintf("%s_%s_%s_%s", idempotencyKey, language, terminalID, audioHash)
	idempotencyHash := "idemp_transcribe_" + utils.HashString(hashInput)

	var existingTaskID string
	if _, exists, _ := uc.cache.GetWithTTL(idempotencyHash, &existingTaskID); exists && existingTaskID != "" {
		status, ok := uc.store.Get(existingTaskID)
		if !ok || status == nil {
			var cachedStatus speechdtos.AsyncTranscriptionStatusDTO
			if _, cachedExists, _ := uc.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
				status = &cachedStatus
				ok = true
			}
		}
		if ok && status != nil && status.Status != "failed" {
			return existingTaskID, true
		}
	}
	return "", false
}

func (uc *transcribeUseCase) TranscribeAudio(ctx context.Context, inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcribe: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	var meta *TranscriptionMetadata
	if len(metadata) > 0 {
		meta = &metadata[0]
	}

	// 1. Idempotency Check
	var idempotencyHash string
	if meta != nil && meta.IdempotencyKey != "" {
		// Create a deterministic hash based on idempotency key, language, terminal ID, and audio content
		audioHash, _ := utils.HashFile(inputPath)
		hashInput := fmt.Sprintf("%s_%s_%s_%s", meta.IdempotencyKey, language, meta.TerminalID, audioHash)
		idempotencyHash = "idemp_transcribe_" + utils.HashString(hashInput)

		// Check if a task already exists for this idempotency key
		var existingTaskID string
		if _, exists, _ := uc.cache.GetWithTTL(idempotencyHash, &existingTaskID); exists && existingTaskID != "" {
			// Check task state - only return if NOT failed
			status, ok := uc.store.Get(existingTaskID)
			if !ok || status == nil {
				// Fallback to cache if memory store is empty
				var cachedStatus speechdtos.AsyncTranscriptionStatusDTO
				if _, cachedExists, _ := uc.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
					status = &cachedStatus
					uc.store.Set(existingTaskID, status)
					ok = true
				}
			}

			if ok && status != nil && status.Status != "failed" {
				utils.LogInfo("Transcribe Task: Duplicate request detected for IdempotencyKey %s. Returning existing TaskID %s (Status: %s)", meta.IdempotencyKey, existingTaskID, status.Status)
				return existingTaskID, nil
			}
			utils.LogInfo("Transcribe Task: Found existing task %s for key %s but it failed or could not be loaded. Proceeding with new task.", existingTaskID, meta.IdempotencyKey)
		}
	}

	taskID := utils.GenerateUUID()

	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		Trigger:   "",
		StartedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if meta != nil {
		if meta.Trigger != "" {
			status.Trigger = meta.Trigger
		}
		if meta.TerminalID != "" {
			status.TerminalID = meta.TerminalID
		}
	}

	// 2. Mark as pending and save idempotency key map
	uc.store.Set(taskID, status)
	_ = uc.cache.Set(taskID, status)

	if idempotencyHash != "" {
		// Store the mapping from idempotency hash to task ID
		_ = uc.cache.Set(idempotencyHash, taskID)
	}

	// Mark terminal as actively transcribing if applicable
	if meta != nil && meta.TerminalID != "" && meta.Source == "mqtt" {
		utils.ActiveTranscriptions.Store(meta.TerminalID, true)
	}

	go uc.processAsync(ctx, taskID, inputPath, language, meta)

	return taskID, nil
}

func (uc *transcribeUseCase) TranscribeAudioSync(ctx context.Context, inputPath string, reqLanguage string, diarize bool, refine bool) (*speechdtos.AsyncTranscriptionResultDTO, error) {
	result, err := uc.whisperClient.Transcribe(ctx, inputPath, reqLanguage, diarize)
	if err != nil && uc.fallbackClient != nil {
		utils.LogWarn("TranscribeSync: Primary client failed, falling back to local: %v", err)
		result, err = uc.fallbackClient.Transcribe(ctx, inputPath, reqLanguage, diarize)
	}

	if err != nil {
		return nil, err
	}

	// Refine (Grammar/Spelling) - only if explicitly requested
	refined := result.Transcription
	if refine {
		refineLang := result.DetectedLanguage
		if reqLanguage != "" {
			refineLang = reqLanguage
		}
		refined, _ = uc.refineUC.RefineText(ctx, result.Transcription, refineLang)
	}

	return &speechdtos.AsyncTranscriptionResultDTO{
		Transcription:    result.Transcription,
		RefinedText:      refined,
		DetectedLanguage: result.DetectedLanguage,
	}, nil
}

func (uc *transcribeUseCase) processAsync(ctx context.Context, taskID string, inputPath string, reqLanguage string, metadata *TranscriptionMetadata) {
	defer func() {
		if metadata != nil && metadata.TerminalID != "" && metadata.Source == "mqtt" {
			utils.ActiveTranscriptions.Delete(metadata.TerminalID)
		}
		if r := recover(); r != nil {
			utils.LogError("Transcribe Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", nil, fmt.Errorf("internal panic: %v", r))
		}
	}()

	diarize := false
	refine := true // Default to true for existing flows
	if metadata != nil {
		diarize = metadata.Diarize
		// If we add Refine to metadata in the future, we should use it here.
	}

	finalResult, err := uc.TranscribeAudioSync(ctx, inputPath, reqLanguage, diarize, refine)
	if err != nil {
		utils.LogError("Transcribe Task %s: Failed: %v", taskID, err)
		uc.updateStatus(taskID, "failed", nil, err)

		if metadata != nil && metadata.DeleteAfter {
			_ = os.Remove(inputPath)
		}
		return
	}

	utils.LogInfo("Transcribe Task %s: Finished successfully", taskID)

	if metadata != nil && metadata.DeleteAfter {
		_ = os.Remove(inputPath)
	}

	uc.updateStatus(taskID, "completed", finalResult, nil)

	// Chaining to /chat ONLY if initiated via MQTT
	if metadata != nil && metadata.Source == "mqtt" && metadata.TerminalID != "" && uc.mqttSvc != nil {
		chatTopic := fmt.Sprintf("users/%s/%s/chat", metadata.TerminalID, uc.config.ApplicationEnvironment)
		prompt := finalResult.RefinedText
		if prompt == "" {
			prompt = finalResult.Transcription
		}

		chatReq := map[string]string{
			"prompt":      prompt,
			"terminal_id": metadata.TerminalID,
			"language":    finalResult.DetectedLanguage,
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
	// Try to get existing status to preserve StartedAt and TerminalID
	var existing speechdtos.AsyncTranscriptionStatusDTO
	_, _, _ = uc.cache.GetWithTTL(taskID, &existing)

	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:     statusStr,
		Result:     result,
		StartedAt:  existing.StartedAt,
		Trigger:    existing.Trigger,
		TerminalID: existing.TerminalID,
		ExpiresAt:  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
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

		// Send MQTT "stop" signal if TerminalID is available
		if status.TerminalID != "" && uc.mqttSvc != nil {
			taskTopic := fmt.Sprintf("users/%s/%s/task", status.TerminalID, uc.config.ApplicationEnvironment)
			msg := map[string]string{
				"event": "stop",
				"task":  "Transcribe",
			}
			payload, _ := json.Marshal(msg)
			if err := uc.mqttSvc.Publish(taskTopic, 0, false, payload); err != nil {
				utils.LogError("Transcribe Task %s: Failed to publish stop signal to MQTT: %v", taskID, err)
			} else {
				utils.LogInfo("Transcribe Task %s: Published stop signal to %s", taskID, taskTopic)
			}
		}
	}

	uc.store.Set(taskID, status)
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
