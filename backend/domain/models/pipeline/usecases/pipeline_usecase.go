package usecases

import (
	"context"
	"fmt"
	"os"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineDtos "sensio/domain/models/pipeline/dtos"
	ragUsecases "sensio/domain/models/rag/usecases"
	speechUsecases "sensio/domain/models/whisper/usecases"
	"strings"
	"time"

	"encoding/json"
	"sensio/domain/common/events"

	"github.com/google/uuid"
)

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

type PipelineUseCase interface {
	ExecutePipeline(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string) (string, error)
	ExecutePipelineWithSession(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string, sessionID string) (string, error)
	CheckIdempotency(idempotencyKey string, audioHash string, req pipelineDtos.PipelineRequestDTO) (string, bool)
}

type pipelineUseCase struct {
	transcribeUC speechUsecases.TranscribeUseCase
	translateUC  ragUsecases.TranslateUseCase
	summaryUC    ragUsecases.SummaryUseCase
	cache        *tasks.BadgerTaskCache
	store        *tasks.StatusStore[pipelineDtos.PipelineStatusDTO]
	mqttSvc      mqttPublisher
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
		transcribeUC: transcribeUC,
		translateUC:  translateUC,
		summaryUC:    summaryUC,
		cache:        cache,
		store:        store,
		mqttSvc:      mqttSvc,
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
	go func() {
		defer cancel()
		u.runPipelineAsync(asyncCtx, taskID, inputPath, req)
	}()

	return taskID, nil
}

func (u *pipelineUseCase) runPipelineAsync(ctx context.Context, taskID string, inputPath string, req pipelineDtos.PipelineRequestDTO) {
	defer os.Remove(inputPath)

	status, _ := u.store.Get(taskID)
	if status == nil {
		return
	}
	status.OverallStatus = "processing"
	u.saveStatus(taskID, *status)

	u.publishEvent(taskID, req.MacAddress, "started", "processing", "", "", 0, nil)

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
		Language:         req.Language,
		Diarize:          req.Diarize,
		Refine:           refine,
		IsPipeline:       true,
		DisableFallback:  true, // MEETING SUMMARY REQUIREMENT: No local fallback for pipeline path
		ProgressCallback: func(progress int) {
			u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "transcription", "processing", progress, nil)
		},
		TerminalContext: []string{req.MacAddress},
	}

	transResult, err := u.transcribeUC.TranscribeAudioSync(ctx, inputPath, transOpts)
	if err != nil {
		u.failStage(taskID, req.MacAddress, "transcription", err)
		return
	}
	status.Stages["transcription"] = pipelineDtos.PipelineStageStatus{
		Status:          "completed",
		Result:          transResult.Transcription,
		DurationSeconds: time.Since(startTime).Seconds(),
	}
	u.saveStatus(taskID, *status)
	u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "transcription", "completed", 100, nil)

	// Stage 2: Refinement
	refinedText := transResult.Transcription
	if status.Stages["refinement"].Status != "skipped" {
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
		startTime = time.Now()
		status.Stages["translation"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "translation", "processing", 0, nil)

		transText, err := u.translateUC.TranslateTextSync(ctx, refinedText, req.TargetLanguage, req.MacAddress)
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
		startTime = time.Now()
		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "summary", "processing", 0, nil)

		participantsStr := strings.Join(req.Participants, ", ")
		summResult, err := u.summaryUC.SummarizeTextSync(ctx, finalText, req.TargetLanguage, req.Context, req.Style, req.Date, req.Location, participantsStr, req.MacAddress)
		if err != nil {
			u.failStage(taskID, req.MacAddress, "summary", err)
			return
		}
		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          summResult,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		utils.LogInfo("Pipeline Task %s: Summary stage completed (Duration: %.2fs)", taskID, status.Stages["summary"].DurationSeconds)
		u.saveStatus(taskID, *status)
		u.publishEvent(taskID, req.MacAddress, "stage_update", "processing", "summary", "completed", 100, nil)
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
	ttl, err := time.ParseDuration(utils.AppConfig.TaskStatusTTL)
	if err != nil {
		ttl = 24 * time.Hour
	}
	u.store.Set(taskID, &status)
	_ = u.cache.SetWithTTL(taskID, status, ttl)
}

func (u *pipelineUseCase) failStage(taskID string, macAddress string, stageName string, err error) {
	status, _ := u.store.Get(taskID)
	if status == nil {
		return
	}
	stage := status.Stages[stageName]
	stage.Status = "failed"
	stage.Error = err.Error()
	status.Stages[stageName] = stage
	status.OverallStatus = "failed"
	u.saveStatus(taskID, *status)

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
