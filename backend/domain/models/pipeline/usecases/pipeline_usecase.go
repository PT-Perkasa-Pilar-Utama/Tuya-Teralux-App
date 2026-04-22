package usecases

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineDtos "sensio/domain/models/pipeline/dtos"
	ragUsecases "sensio/domain/models/rag/usecases"
	speechUsecases "sensio/domain/models/whisper/usecases"
	"strings"
	"sync"
	"time"

	"encoding/json"
	"sensio/domain/common/events"

	"github.com/google/uuid"
)

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// PipelineUseCase orchestrates the 4-stage meeting processing pipeline:
// Transcription -> Refinement -> Translation -> Summary
//
// ARCHITECTURE NOTE: Pipeline results are stored in in-memory/cache status stores
// with TTL (default 24h). Structured artifacts (utterances, segments, action items,
// decisions, etc.) are EPHEMERAL and will be lost after TTL expiry or server restart.
//
// For persistent storage of structured meeting artifacts, future work should:
// 1. Add tables for transcription_segments, meeting_decisions, action_items, etc.
// 2. Persist structured results after each stage completion
// 3. Implement retrieval APIs for historical meeting data
//
// Current persistence: Only recording metadata is persisted to DB
// (see: domain/recordings/entities/recording.go)
type PipelineUseCase interface {
	ExecutePipeline(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string) (string, error)
	ExecutePipelineWithSession(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string, sessionID string) (string, error)
	CheckIdempotency(idempotencyKey string, audioHash string, req pipelineDtos.PipelineRequestDTO) (string, bool)
	CancelTask(taskID string) error
}

type pipelineUseCase struct {
	transcribeUC   speechUsecases.TranscribeUseCase
	translateUC    ragUsecases.TranslateUseCase
	summaryUC      ragUsecases.SummaryUseCase
	cache          *tasks.BadgerTaskCache
	store          *tasks.StatusStore[pipelineDtos.PipelineStatusDTO]
	mqttSvc        mqttPublisher
	cancelRegistry map[string]context.CancelFunc
	registryMu     sync.RWMutex
}

func NewPipelineUseCase(
	transcribeUC speechUsecases.TranscribeUseCase,
	translateUC ragUsecases.TranslateUseCase,
	summaryUC ragUsecases.SummaryUseCase,
	cache *tasks.BadgerTaskCache,
	store *tasks.StatusStore[pipelineDtos.PipelineStatusDTO],
	mqttSvc mqttPublisher,
) PipelineUseCase {
	return &pipelineUseCase{
		transcribeUC:   transcribeUC,
		translateUC:    translateUC,
		summaryUC:      summaryUC,
		cache:          cache,
		store:          store,
		mqttSvc:        mqttSvc,
		cancelRegistry: make(map[string]context.CancelFunc),
	}
}

func (u *pipelineUseCase) CheckIdempotency(idempotencyKey string, audioHash string, req pipelineDtos.PipelineRequestDTO) (string, bool) {
	if idempotencyKey == "" {
		return "", false
	}
	hashInput := fmt.Sprintf("%s_%s_%s_%s_%s", idempotencyKey, req.Language, req.TargetLanguage, req.MacAddress, audioHash)
	idempotencyHash := "idemp_pipeline_" + utils.HashString(hashInput)

	var existingTaskID string
	_, exists, _ := u.cache.GetWithTTL(idempotencyHash, &existingTaskID)
	if exists && existingTaskID != "" {
		existingStatus, _ := u.store.Get(existingTaskID)
		if existingStatus == nil {
			var cachedStatus pipelineDtos.PipelineStatusDTO
			if _, cachedExists, _ := u.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
				existingStatus = &cachedStatus
			}
		}
		if existingStatus != nil && existingStatus.OverallStatus != "failed" {
			return existingTaskID, true
		}
	}
	return "", false
}

func (u *pipelineUseCase) ExecutePipeline(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string) (string, error) {
	return u.ExecutePipelineWithSession(ctx, inputPath, req, idempotencyKey, "")
}

// ExecutePipelineWithSession executes a pipeline with optional session ID for by-upload requests.
// When sessionID is provided, it is included in the idempotency hash to prevent cross-session collisions.
func (u *pipelineUseCase) ExecutePipelineWithSession(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string, sessionID string) (string, error) {
	// 1. Idempotency Check
	var idempotencyHash string
	if idempotencyKey != "" {
		audioHash, _ := utils.HashFile(inputPath)
		// For by-upload requests, include session ID in hash to prevent cross-session collisions
		var hashInput string
		if sessionID != "" {
			hashInput = fmt.Sprintf("%s_%s_%s_%s_%s_session:%s", idempotencyKey, req.Language, req.TargetLanguage, req.MacAddress, audioHash, sessionID)
		} else {
			hashInput = fmt.Sprintf("%s_%s_%s_%s_%s", idempotencyKey, req.Language, req.TargetLanguage, req.MacAddress, audioHash)
		}
		idempotencyHash = "idemp_pipeline_" + utils.HashString(hashInput)

		var existingTaskID string
		_, exists, _ := u.cache.GetWithTTL(idempotencyHash, &existingTaskID)
		if exists && existingTaskID != "" {
			// Check if existing task is failed. If so, we allow retry.
			existingStatus, _ := u.store.Get(existingTaskID)
			if existingStatus == nil {
				// Fallback to cache if memory store is empty (e.g. after restart)
				var cachedStatus pipelineDtos.PipelineStatusDTO
				if _, cachedExists, _ := u.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
					existingStatus = &cachedStatus
					// Prime the store for future calls
					u.store.Set(existingTaskID, existingStatus)
				}
			}

			if existingStatus != nil && existingStatus.OverallStatus != "failed" {
				utils.LogInfo("Pipeline: Duplicate request detected for IdempotencyKey %s (session: %s). Returning existing TaskID %s", idempotencyKey, sessionID, existingTaskID)
				return existingTaskID, nil
			}
			utils.LogInfo("Pipeline: Existing task %s found for IdempotencyKey %s (session: %s) but it is in 'failed' state or could not be loaded. Starting new task.", existingTaskID, idempotencyKey, sessionID)
		}

		// 2. Atomic Idempotency Check & Set
		// We use a temporary lock key to ensure only one task is created for this hash
		lockKey := "lock:" + idempotencyHash
		isNew, err := u.cache.SetIfAbsentWithTTL(lockKey, []byte("1"), 30*time.Second)
		if err != nil {
			utils.LogError("Pipeline: Failed to acquire idempotency lock | hash=%s | error=%v", idempotencyHash, err)
		} else if !isNew {
			// Lock exists, meaning another request is currently creating the task or it was just created
			// Wait briefly and check if the actual task ID has been set
			for i := 0; i < 5; i++ {
				time.Sleep(200 * time.Millisecond)
				var currentTaskID string
				_, exists, _ := u.cache.GetWithTTL(idempotencyHash, &currentTaskID)
				if exists && currentTaskID != "" && currentTaskID != existingTaskID {
					currentStatus, _ := u.store.Get(currentTaskID)
					if currentStatus == nil {
						var cachedStatus pipelineDtos.PipelineStatusDTO
						if _, cachedExists, _ := u.cache.GetWithTTL(currentTaskID, &cachedStatus); cachedExists {
							currentStatus = &cachedStatus
						}
					}
					if currentStatus != nil && currentStatus.OverallStatus != "failed" {
						utils.LogInfo("Pipeline: Duplicate request detected via lock | hash=%s | taskID=%s", idempotencyHash, currentTaskID)
						return currentTaskID, nil
					}
				}
			}
			utils.LogWarn("Pipeline: Lock acquired by another process but task was not created or failed. Proceeding with caution.")
		}
		// Ensure we release the lock if we created it but later decided not to proceed or finished
		defer u.cache.Delete(lockKey)
	}

	taskID := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	status := pipelineDtos.PipelineStatusDTO{
		TaskID:        taskID,
		OverallStatus: "pending",
		StartedAt:     now,
		MacAddress:    req.MacAddress,
		Stages: map[string]pipelineDtos.PipelineStageStatus{
			"transcription": {Status: "pending"},
			"refinement":    {Status: "pending"},
			"translation":   {Status: "pending"},
			"summary":       {Status: "pending"},
		},
	}

	// Optimization: Skip stages if not requested or not needed
	if req.Refine != nil && !*req.Refine {
		status.Stages["refinement"] = pipelineDtos.PipelineStageStatus{Status: "skipped"}
	}
	if req.Language == req.TargetLanguage || req.TargetLanguage == "" {
		status.Stages["translation"] = pipelineDtos.PipelineStageStatus{Status: "skipped"}
	}
	if !req.Summarize {
		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{Status: "skipped"}
	}

	u.saveStatus(taskID, status)
	if idempotencyHash != "" {
		_ = u.cache.Set(idempotencyHash, taskID)
	}

	timeout, err := time.ParseDuration(utils.AppConfig.PipelineAsyncTimeout)
	if err != nil {
		timeout = 12 * time.Hour
	}

	u.publishEvent(taskID, req.MacAddress, "accepted", "pending", "", "", 0, nil)

	asyncCtx, cancel := context.WithTimeout(context.Background(), timeout)
	u.registerCancel(taskID, cancel)
	go func() {
		defer cancel()
		defer u.unregisterCancel(taskID)
		u.runPipelineAsync(asyncCtx, taskID, inputPath, req)
	}()

	return taskID, nil
}

func (u *pipelineUseCase) runPipelineAsync(ctx context.Context, taskID string, inputPath string, req pipelineDtos.PipelineRequestDTO) {
	// Determine if input audio should be preserved (meeting-summary jobs)
	preserveInputAudio := req.Summarize
	if preserveInputAudio {
		utils.LogInfo("Pipeline Task %s: Preserving input audio (meeting summary) | path=%s", taskID, inputPath)
	} else {
		defer os.Remove(inputPath)
		utils.LogInfo("Pipeline Task %s: Will delete input audio after pipeline | path=%s", taskID, inputPath)
	}

	status, _ := u.store.Get(taskID)
	if status == nil {
		return
	}
	status.OverallStatus = "processing"
	u.saveStatus(taskID, *status)

	u.publishEvent(taskID, req.MacAddress, "started", "processing", "", "", 0, nil)

	// Check for cancellation before starting Stage 1
	select {
	case <-ctx.Done():
		u.handleCancellation(taskID, "", status, req.MacAddress)
		return
	default:
	}

	// Stage 1: Transcription
	startTime := time.Now()
	status.Stages["transcription"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
	u.saveStatus(taskID, *status)
	u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "transcription", "processing", 0, nil)

	refine := true
	if req.Refine != nil && !*req.Refine {
		refine = false
	}

	transOpts := speechUsecases.TranscribeOptions{
		Language:        req.Language,
		Diarize:         req.Diarize,
		Refine:          refine,
		IsPipeline:      true,
		DisableFallback: true, // MEETING SUMMARY REQUIREMENT: No local fallback for pipeline path
		ProgressCallback: func(progress int) {
			u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "transcription", "processing", progress, nil)
		},
		TerminalContext: []string{req.MacAddress},
	}

	transResult, err := u.transcribeUC.TranscribeAudioSync(ctx, inputPath, transOpts)
	// Check for cancellation immediately after blocking call
	select {
	case <-ctx.Done():
		u.handleCancellation(taskID, "transcription", status, req.MacAddress)
		return
	default:
	}

	if err != nil {
		u.failStage(taskID, req.MacAddress, "transcription", err)
		return
	}

	// OBSERVABILITY: Log transcription quality indicators
	utils.LogInfo("Pipeline Task %s: Transcription completed (Duration: %.2fs)", taskID, time.Since(startTime).Seconds())

	if transResult.TranscriptFormat != "" {
		utils.LogInfo("Pipeline Task %s: Transcript format: %s", taskID, transResult.TranscriptFormat)
	}
	if len(transResult.Utterances) > 0 {
		utils.LogInfo("Pipeline Task %s: Extracted %d utterances with speaker diarization", taskID, len(transResult.Utterances))
	} else if req.Diarize {
		utils.LogWarn("Pipeline Task %s: WARNING - Diarization requested but no utterances extracted (provider may not support it)", taskID)
	}
	if len(transResult.Segments) > 0 {
		utils.LogInfo("Pipeline Task %s: Segmented transcription with %d segments", taskID, len(transResult.Segments))
	}
	if transResult.ConfidenceSummary != nil {
		utils.LogInfo("Pipeline Task %s: Confidence avg: %.2f (utterances: %d)",
			taskID, transResult.ConfidenceSummary.AverageConfidence,
			transResult.ConfidenceSummary.UtterancesCount)
	}

	status.Stages["transcription"] = pipelineDtos.PipelineStageStatus{
		Status:          "completed",
		Result:          transResult, // Store full result with utterances, segments, etc.
		DurationSeconds: time.Since(startTime).Seconds(),
	}
	u.saveStatus(taskID, *status)
	u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "transcription", "completed", 100, nil)

	// Stage 2: Refinement
	refinedText := transResult.Transcription
	if status.Stages["refinement"].Status != "skipped" {
		// Check for cancellation before starting refinement
		select {
		case <-ctx.Done():
			u.handleCancellation(taskID, "transcription", status, req.MacAddress)
			return
		default:
		}

		startTime = time.Now()
		status.Stages["refinement"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "refinement", "processing", 0, nil)

		// Refine is already called inside TranscribeAudioSync and returned as transResult.RefinedText
		// But if we want to be explicit or if we separated them:
		refinedText = transResult.RefinedText
		status.Stages["refinement"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          refinedText,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "refinement", "completed", 100, nil)
	}

	// Stage 3: Translation
	finalText := refinedText
	if status.Stages["translation"].Status != "skipped" {
		// Check for cancellation before starting translation
		select {
		case <-ctx.Done():
			u.handleCancellation(taskID, "refinement", status, req.MacAddress)
			return
		default:
		}

		startTime = time.Now()
		status.Stages["translation"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "translation", "processing", 0, nil)

		transText, err := u.translateUC.TranslateTextSync(ctx, refinedText, req.TargetLanguage, req.MacAddress)
		// Check for cancellation immediately after blocking call
		select {
		case <-ctx.Done():
			u.handleCancellation(taskID, "translation", status, req.MacAddress)
			return
		default:
		}

		if err != nil {
			u.failStage(taskID, req.MacAddress, "translation", err)
			return
		}
		finalText = transText
		status.Stages["translation"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          finalText,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "translation", "completed", 100, nil)
	}

	// Stage 4: Summary
	if status.Stages["summary"].Status != "skipped" {
		// Check for cancellation before starting summary
		select {
		case <-ctx.Done():
			u.handleCancellation(taskID, "translation", status, req.MacAddress)
			return
		default:
		}

		startTime = time.Now()
		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "summary", "processing", 0, nil)

		participantsStr := strings.Join(req.Participants, ", ")
		summResult, err := u.summaryUC.SummarizeTextSync(ctx, finalText, req.TargetLanguage, req.Context, req.Style, req.Date, req.Location, participantsStr, req.MacAddress)
		// Check for cancellation immediately after blocking call
		select {
		case <-ctx.Done():
			u.handleCancellation(taskID, "summary", status, req.MacAddress)
			return
		default:
		}

		if err != nil {
			u.failStage(taskID, req.MacAddress, "summary", err)
			return
		}

		// OBSERVABILITY: Log summary mode and structured artifacts
		summaryMode := "single_pass"
		if summResult.SummaryMode != "" {
			summaryMode = summResult.SummaryMode
		}

		utils.LogInfo("Pipeline Task %s: Summary stage completed (Duration: %.2fs, Mode: %s)",
			taskID, time.Since(startTime).Seconds(), summaryMode)

		// Log structured artifact counts if available
		if len(summResult.ActionItems) > 0 {
			utils.LogInfo("Pipeline Task %s: Extracted %d action items", taskID, len(summResult.ActionItems))
		}
		if len(summResult.Decisions) > 0 {
			utils.LogInfo("Pipeline Task %s: Extracted %d decisions", taskID, len(summResult.Decisions))
		}
		if len(summResult.OpenIssues) > 0 {
			utils.LogInfo("Pipeline Task %s: Extracted %d open issues", taskID, len(summResult.OpenIssues))
		}
		if len(summResult.Risks) > 0 {
			utils.LogInfo("Pipeline Task %s: Extracted %d risks", taskID, len(summResult.Risks))
		}
		if summResult.CoverageStats != nil {
			utils.LogInfo("Pipeline Task %s: Coverage ratio: %.2f (windows: %d/%d)",
				taskID, summResult.CoverageStats.CoverageRatio,
				summResult.CoverageStats.ProcessedWindows,
				summResult.CoverageStats.TotalWindows)
		}

		// Warning logs for risky degradations
		if summResult.SummaryMode == "single_pass" && len(finalText) > 16000 {
			utils.LogWarn("Pipeline Task %s: WARNING - Long transcript (%d chars) used single_pass mode (hierarchical failed or not triggered)", taskID, len(finalText))
		}

		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          summResult,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "summary", "completed", 100, nil)
	}

	// Final cancellation check before marking task as completed
	// This prevents race condition where CancelTask() is called after summary completes
	// but before overall_status is set to completed
	select {
	case <-ctx.Done():
		u.handleCancellation(taskID, "summary", status, req.MacAddress)
		return
	default:
	}

	// Finalize
	status.OverallStatus = "completed"
	start, _ := time.Parse(time.RFC3339, status.StartedAt)
	duration := time.Since(start).Seconds()
	status.DurationSeconds = duration
	u.saveStatus(taskID, *status)

	u.publishEvent(taskID, req.MacAddress, "completed", "completed", "", "", 100, nil)

	utils.LogInfo("Pipeline Task %s: completed (Duration: %.2fs)", taskID, duration)
}

func (u *pipelineUseCase) saveStatus(taskID string, status pipelineDtos.PipelineStatusDTO) {
	mu := u.store.GetTaskMutex(taskID)
	mu.Lock()
	defer mu.Unlock()

	ttl, err := time.ParseDuration(utils.AppConfig.TaskStatusTTL)
	if err != nil {
		ttl = 24 * time.Hour
	}
	status.Version++
	u.store.Set(taskID, &status)
	_ = u.cache.SetWithTTL(taskID, status, ttl)
}

func (u *pipelineUseCase) failStage(taskID string, macAddress string, stageName string, err error) {
	mu := u.store.GetTaskMutex(taskID)
	mu.Lock()
	defer mu.Unlock()

	status, _ := u.store.Get(taskID)
	if status == nil {
		return
	}
	stage := status.Stages[stageName]
	stage.Status = "failed"
	stage.Error = err.Error()
	status.Stages[stageName] = stage
	status.OverallStatus = "failed"
	status.Version++

	ttl, _ := time.ParseDuration(utils.AppConfig.TaskStatusTTL)
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	u.store.Set(taskID, status)
	_ = u.cache.SetWithTTL(taskID, *status, ttl)

	utils.LogError("Pipeline Task %s: Stage '%s' failed: %v", taskID, stageName, err)
	u.publishEvent(taskID, macAddress, "failed", "failed", stageName, "failed", 0, err)
}

// publishEvent publishes TaskEventV1 to MQTT
func (u *pipelineUseCase) publishEvent(taskID string, macAddress string, event string, overallStatus string, stage string, stageStatus string, progress int, err error) {
	if !utils.GetConfig().TaskEventPublishEnabled || u.mqttSvc == nil || macAddress == "" {
		return
	}

	taskEvent := events.NewTaskEventV1(taskID, "MeetingPipeline", event, overallStatus)
	if stage != "" {
		taskEvent.Stage = stage
		taskEvent.StageStatus = stageStatus
	}
	if progress > 0 {
		taskEvent.ProgressPercent = progress
	}
	if err != nil {
		taskEvent.Error = err.Error()
	}

	topic := fmt.Sprintf("users/%s/%s/task", macAddress, utils.GetConfig().ApplicationEnvironment)
	payloadBytes, _ := json.Marshal(taskEvent)
	_ = u.mqttSvc.Publish(topic, 0, false, payloadBytes)
}

// registerCancel stores the cancel function for a task
func (u *pipelineUseCase) registerCancel(taskID string, cancel context.CancelFunc) {
	u.registryMu.Lock()
	defer u.registryMu.Unlock()
	u.cancelRegistry[taskID] = cancel
}

// unregisterCancel removes the cancel function for a task
func (u *pipelineUseCase) unregisterCancel(taskID string) {
	u.registryMu.Lock()
	defer u.registryMu.Unlock()
	delete(u.cancelRegistry, taskID)
}

// getAndRemoveCancel retrieves and removes the cancel function for a task
func (u *pipelineUseCase) getAndRemoveCancel(taskID string) context.CancelFunc {
	u.registryMu.Lock()
	defer u.registryMu.Unlock()
	cancel := u.cancelRegistry[taskID]
	delete(u.cancelRegistry, taskID)
	return cancel
}

// CancelTask cancels an active pipeline task
func (u *pipelineUseCase) CancelTask(taskID string) error {
	status, found := u.store.Get(taskID)
	if !found || status == nil {
		// Also check cache in case store doesn't have it
		var cachedStatus pipelineDtos.PipelineStatusDTO
		_, cachedExists, _ := u.cache.GetWithTTL(taskID, &cachedStatus)
		if !cachedExists {
			return errors.New("task not found")
		}
		status = &cachedStatus
	}

	// Check if task is already in a terminal state
	if status.OverallStatus == "completed" || status.OverallStatus == "cancelled" || status.OverallStatus == "failed" {
		// Already terminal - treat as no-op success
		utils.LogInfo("Pipeline: CancelTask called for task %s already in terminal state: %s", taskID, status.OverallStatus)
		return nil
	}

	// Trigger cancellation
	cancel := u.getAndRemoveCancel(taskID)
	if cancel != nil {
		cancel()
	}

	// Mark task as cancelled
	u.markCancelled(taskID, status)

	utils.LogInfo("Pipeline: Task %s cancelled successfully", taskID)
	return nil
}

// handleCancellation handles cooperative cancellation during pipeline execution
func (u *pipelineUseCase) handleCancellation(taskID string, currentStage string, status *pipelineDtos.PipelineStatusDTO, macAddress string) {
	utils.LogInfo("Pipeline: Task %s cancelled during stage: %s", taskID, currentStage)

	// Mark the current stage as cancelled if we know which stage was active
	if currentStage != "" {
		stage := status.Stages[currentStage]
		stage.Status = "cancelled"
		stage.Error = "task cancelled by user"
		status.Stages[currentStage] = stage
	}

	// Set overall status to cancelled
	status.OverallStatus = "cancelled"
	u.saveStatus(taskID, *status)

	// Publish MQTT cancellation event
	u.publishCancelledEvent(taskID, currentStage, macAddress)
}

// markCancelled marks a task as cancelled and publishes MQTT event
func (u *pipelineUseCase) markCancelled(taskID string, status *pipelineDtos.PipelineStatusDTO) {
	// Find the active stage (the one that was processing)
	activeStage := ""
	for stageName, stageStatus := range status.Stages {
		if stageStatus.Status == "processing" {
			activeStage = stageName
			break
		}
	}

	// Set the active stage to cancelled
	if activeStage != "" {
		stage := status.Stages[activeStage]
		stage.Status = "cancelled"
		stage.Error = "task cancelled by user"
		status.Stages[activeStage] = stage
	}

	// Set overall status to cancelled
	status.OverallStatus = "cancelled"
	u.saveStatus(taskID, *status)

	// Publish MQTT cancellation event
	u.publishCancelledEvent(taskID, activeStage, status.MacAddress)
}

// publishCancelledEvent publishes a cancellation event to MQTT
func (u *pipelineUseCase) publishCancelledEvent(taskID string, stage string, macAddress string) {
	if !utils.GetConfig().TaskEventPublishEnabled || u.mqttSvc == nil || macAddress == "" {
		return
	}

	taskEvent := events.NewTaskEventV1(taskID, "MeetingPipeline", "cancelled", "cancelled")
	if stage != "" {
		taskEvent.Stage = stage
		taskEvent.StageStatus = "cancelled"
	}
	taskEvent.Error = "task cancelled by user"

	topic := fmt.Sprintf("users/%s/%s/task", macAddress, utils.GetConfig().ApplicationEnvironment)
	payloadBytes, _ := json.Marshal(taskEvent)
	_ = u.mqttSvc.Publish(topic, 0, false, payloadBytes)

	utils.LogInfo("Pipeline: Published cancellation event for task %s on topic %s", taskID, topic)
}
