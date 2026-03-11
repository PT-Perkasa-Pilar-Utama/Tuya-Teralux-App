package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	ragUsecases "sensio/domain/models/rag/usecases"
	whisperdtos "sensio/domain/models/whisper/dtos"
	"strings"
	"sync"
	"time"
)

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// WhisperClient is the unified interface for all whisper transcription services
type WhisperClient interface {
	Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*whisperdtos.WhisperResult, error)
}

type TranscriptionMetadata struct {
	UID            string
	TerminalID     string
	RequestID      string
	Source         string // "mqtt", "rest", etc.
	Trigger        string // e.g., "/api/whisper/transcribe"
	DeleteAfter    bool   // Whether to delete the audio file after processing
	Diarize        bool   // Whether to perform speaker diarization
	IdempotencyKey string // Client-provided idempotency key
}

type TranscribeUseCase interface {
	TranscribeAudio(ctx context.Context, inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error)
	TranscribeAudioSync(ctx context.Context, inputPath string, language string, diarize bool, refine bool, progressCallback func(int)) (*whisperdtos.AsyncTranscriptionResultDTO, error)
	CheckIdempotency(idempotencyKey string, audioHash string, language string, terminalID string) (string, bool)
}

type transcribeUseCase struct {
	whisperClient  WhisperClient
	fallbackClient WhisperClient
	refineUC       ragUsecases.RefineUseCase
	store          *tasks.StatusStore[whisperdtos.AsyncTranscriptionStatusDTO]
	cache          *tasks.BadgerTaskCache
	config         *utils.Config
	mqttSvc        mqttPublisher
}

func NewTranscribeUseCase(
	whisperClient WhisperClient,
	fallbackClient WhisperClient,
	refineUC ragUsecases.RefineUseCase,
	store *tasks.StatusStore[whisperdtos.AsyncTranscriptionStatusDTO],
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
			var cachedStatus whisperdtos.AsyncTranscriptionStatusDTO
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
				var cachedStatus whisperdtos.AsyncTranscriptionStatusDTO
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

	ttl, err := time.ParseDuration(uc.config.TaskStatusTTL)
	if err != nil {
		ttl = 24 * time.Hour
	}

	status := &whisperdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		Trigger:   "",
		StartedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(ttl).Format(time.RFC3339),
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

	timeout, err := time.ParseDuration(uc.config.TranscribeAsyncTimeout)
	if err != nil {
		timeout = 8 * time.Hour
	}
	asyncCtx, cancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		defer cancel()
		uc.processAsync(asyncCtx, taskID, inputPath, language, meta)
	}()

	return taskID, nil
}

func (uc *transcribeUseCase) TranscribeAudioSync(ctx context.Context, inputPath string, reqLanguage string, diarize bool, refine bool, progressCallback func(int)) (*whisperdtos.AsyncTranscriptionResultDTO, error) {
	var rawTranscription string
	var detectedLang string

	// 1. Audio Normalization (WAV PCM 16k Mono)
	processingPath, audioCleanup, err := utils.NormalizeToWavPCM16k(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize audio: %w", err)
	}
	defer audioCleanup()

	fileInfo, err := os.Stat(processingPath)
	useSegment := err == nil && fileInfo.Size() > 20*1024*1024 && uc.config.AudioSegmentEnabled

	if useSegment {
		segmentSec := uc.config.AudioSegmentSec
		if segmentSec < 60 {
			segmentSec = 60
		}
		overlapSec := uc.config.AudioSegmentOverlapSec

		utils.LogInfo("TranscribeSync: Large audio file detected (%d bytes), starting segmented transcription (step=%ds, overlap=%ds)...", fileInfo.Size(), segmentSec, overlapSec)
		segments, splitErr := utils.SplitAudioSegments(processingPath, segmentSec, overlapSec)
		if splitErr != nil {
			utils.LogError("TranscribeSync: Failed to split audio, falling back to full file: %v", splitErr)
			useSegment = false // Fallback to full file
		} else {
			defer utils.CleanupSegments(segments)

			type segResult struct {
				index int
				text  string
				lang  string
				err   error
			}

			results := make([]segResult, len(segments))
			sem := make(chan struct{}, utils.MaxInt(1, uc.config.AudioSegmentMaxConcurrency))
			var wg sync.WaitGroup

			var completed int
			var lastProgress int
			var mu sync.Mutex

			for i, seg := range segments {
				wg.Add(1)
				go func(idx int, path string) {
					defer wg.Done()

					// Check context before starting
					select {
					case <-ctx.Done():
						mu.Lock()
						results[idx] = segResult{index: idx, err: ctx.Err()}
						mu.Unlock()
						return
					case sem <- struct{}{}:
						defer func() { <-sem }()
					}

					var res *whisperdtos.WhisperResult
					var transErr error

					// Retry logic (max 2 attempts)
					for attempt := 1; attempt <= 2; attempt++ {
						res, transErr = uc.whisperClient.Transcribe(ctx, path, reqLanguage, diarize)
						if transErr != nil && uc.fallbackClient != nil {
							utils.LogWarn("TranscribeSync Segment %d (Attempt %d): Primary client failed, falling back to local: %v", idx, attempt, transErr)
							res, transErr = uc.fallbackClient.Transcribe(ctx, path, reqLanguage, diarize)
						}
						if transErr == nil {
							break
						}
						utils.LogWarn("TranscribeSync Segment %d (Attempt %d) failed: %v", idx, attempt, transErr)
						if attempt < 2 {
							time.Sleep(1 * time.Second)
						}
					}

					mu.Lock()
					defer mu.Unlock()
					if transErr != nil {
						results[idx] = segResult{index: idx, err: transErr}
					} else {
						results[idx] = segResult{index: idx, text: res.Transcription, lang: res.DetectedLanguage}
					}

					completed++
					if progressCallback != nil {
						progress := int((float64(completed) / float64(len(segments))) * 100)
						if progress > lastProgress {
							lastProgress = progress
							go progressCallback(progress)
						}
					}
				}(i, seg.Path)
			}
			wg.Wait()

			// Merge and Dedup
			var mergedText string
			for i, r := range results {
				if r.err != nil {
					return nil, fmt.Errorf("segment %d failed: %w", r.index, r.err)
				}
				if i == 0 {
					mergedText = r.text
					detectedLang = r.lang
				} else {
					mergedText = uc.mergeWithDedup(mergedText, r.text)
					if detectedLang == "" && r.lang != "" {
						detectedLang = r.lang
					}
				}
			}
			rawTranscription = strings.TrimSpace(mergedText)
		}
	}

	if !useSegment {
		result, err := uc.whisperClient.Transcribe(ctx, processingPath, reqLanguage, diarize)
		if err != nil && uc.fallbackClient != nil {
			utils.LogWarn("TranscribeSync: Primary client failed, falling back to local: %v", err)
			result, err = uc.fallbackClient.Transcribe(ctx, processingPath, reqLanguage, diarize)
		}

		if err != nil {
			return nil, err
		}
		rawTranscription = result.Transcription
		detectedLang = result.DetectedLanguage
		if progressCallback != nil {
			go progressCallback(100)
		}
	}

	// Refine (Grammar/Spelling) - only if explicitly requested
	refined := rawTranscription
	if refine {
		refineLang := detectedLang
		if reqLanguage != "" {
			refineLang = reqLanguage
		}
		refined, _ = uc.refineUC.RefineText(ctx, rawTranscription, refineLang)
	}

	return &whisperdtos.AsyncTranscriptionResultDTO{
		Transcription:    rawTranscription,
		RefinedText:      refined,
		DetectedLanguage: detectedLang,
	}, nil
}

func (uc *transcribeUseCase) processAsync(ctx context.Context, taskID string, inputPath string, reqLanguage string, metadata *TranscriptionMetadata) {
	defer func() {
		// Defensive cleanup in case of panic or early exit
		if metadata != nil && metadata.TerminalID != "" && metadata.Source == "mqtt" {
			utils.ActiveTranscriptions.Delete(metadata.TerminalID)
		}
		if r := recover(); r != nil {
			utils.LogError("Transcribe Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", nil, fmt.Errorf("internal panic: %v", r))
		}
	}()

	diarize := false
	refine := false // Default to false to reduce latency
	if metadata != nil {
		diarize = metadata.Diarize
		// If we add Refine to metadata in the future, we should use it here.
	}

	finalResult, err := uc.TranscribeAudioSync(ctx, inputPath, reqLanguage, diarize, refine, nil)
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

	// Explicitly clear the transcription flag BEFORE chaining to /chat
	// to prevent a race condition where the chat controller drops the request.
	if metadata != nil && metadata.TerminalID != "" && metadata.Source == "mqtt" {
		utils.ActiveTranscriptions.Delete(metadata.TerminalID)
	}

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
			"request_id":  metadata.RequestID,
		}
		payload, _ := json.Marshal(chatReq)
		if err := uc.mqttSvc.Publish(chatTopic, 0, false, payload); err != nil {
			utils.LogError("TranscribeUseCase: Failed to publish transcript to MQTT: %v", err)
		}
		utils.LogInfo("Transcribe Task %s: Chained result to %s", taskID, chatTopic)
	}
}

func (uc *transcribeUseCase) updateStatus(taskID string, statusStr string, result *whisperdtos.AsyncTranscriptionResultDTO, err error) {
	// Try to get existing status to preserve StartedAt and TerminalID
	var existing whisperdtos.AsyncTranscriptionStatusDTO
	_, _, _ = uc.cache.GetWithTTL(taskID, &existing)

	ttl, pErr := time.ParseDuration(uc.config.TaskStatusTTL)
	if pErr != nil {
		ttl = 24 * time.Hour
	}

	status := &whisperdtos.AsyncTranscriptionStatusDTO{
		Status:     statusStr,
		Result:     result,
		StartedAt:  existing.StartedAt,
		Trigger:    existing.Trigger,
		TerminalID: existing.TerminalID,
		ExpiresAt:  time.Now().Add(ttl).Format(time.RFC3339),
	}

	if err != nil {
		status.Error = err.Error()
		status.HTTPStatusCode = utils.GetErrorStatusCode(err)
	} else if statusStr == "completed" {
		status.HTTPStatusCode = 200
	}

	// Calculate duration if finished
	if statusStr == "completed" || statusStr == "failed" {
		var duration float64
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			duration = time.Since(startTime).Seconds()
			status.DurationSeconds = duration
		}

		logMsg := fmt.Sprintf("Task %s: %s (Duration: %.2fs)", taskID, statusStr, duration)
		if status.TerminalID != "" {
			logMsg += fmt.Sprintf(", Terminal: %s", status.TerminalID)
		}
		if status.Trigger != "" {
			logMsg += fmt.Sprintf(", Trigger: %s", status.Trigger)
		}
		if err != nil {
			utils.LogError("%s, Error: %v", logMsg, err)
		} else {
			utils.LogInfo("%s", logMsg)
		}
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
	_ = uc.cache.SetWithTTL(taskID, status, ttl)
}

// mergeWithDedup handles the joining of two transcript segments by checking for shared word overlaps
// at the boundaries, which often occurs due to the segment overlap duration.
func (uc *transcribeUseCase) mergeWithDedup(prev, current string) string {
	prevWords := strings.Fields(prev)
	currentWords := strings.Fields(current)

	if len(prevWords) == 0 {
		return current
	}
	if len(currentWords) == 0 {
		return prev
	}

	// Try to find the longest overlap (max 10 words)
	maxOverlap := 10
	if len(prevWords) < maxOverlap {
		maxOverlap = len(prevWords)
	}
	if len(currentWords) < maxOverlap {
		maxOverlap = len(currentWords)
	}

	bestOverlap := 0
	for size := 1; size <= maxOverlap; size++ {
		match := true
		for i := 0; i < size; i++ {
			pWord := prevWords[len(prevWords)-size+i]
			cWord := currentWords[i]
			// Case-insensitive comparison for deduplication
			if !strings.EqualFold(pWord, cWord) {
				match = false
				break
			}
		}
		if match {
			bestOverlap = size
		}
	}

	if bestOverlap > 0 {
		remaining := currentWords[bestOverlap:]
		if len(remaining) == 0 {
			return prev
		}
		return prev + " " + strings.Join(remaining, " ")
	}

	return prev + " " + current
}
