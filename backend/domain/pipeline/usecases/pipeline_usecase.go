package usecases

import (
	"context"
	"fmt"
	"os"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineDtos "sensio/domain/pipeline/dtos"
	ragUsecases "sensio/domain/rag/usecases"
	speechUsecases "sensio/domain/speech/usecases"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PipelineUseCase interface {
	ExecutePipeline(ctx context.Context, inputPath string, req pipelineDtos.PipelineRequestDTO, idempotencyKey string) (string, error)
	CheckIdempotency(idempotencyKey string, audioHash string, req pipelineDtos.PipelineRequestDTO) (string, bool)
}

type pipelineUseCase struct {
	transcribeUC speechUsecases.TranscribeUseCase
	translateUC  ragUsecases.TranslateUseCase
	summaryUC    ragUsecases.SummaryUseCase
	cache        *tasks.BadgerTaskCache
	store        *tasks.StatusStore[pipelineDtos.PipelineStatusDTO]
}

func NewPipelineUseCase(
	transcribeUC speechUsecases.TranscribeUseCase,
	translateUC ragUsecases.TranslateUseCase,
	summaryUC ragUsecases.SummaryUseCase,
	cache *tasks.BadgerTaskCache,
	store *tasks.StatusStore[pipelineDtos.PipelineStatusDTO],
) PipelineUseCase {
	return &pipelineUseCase{
		transcribeUC: transcribeUC,
		translateUC:  translateUC,
		summaryUC:    summaryUC,
		cache:        cache,
		store:        store,
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
	// 1. Idempotency Check
	var idempotencyHash string
	if idempotencyKey != "" {
		audioHash, _ := utils.HashFile(inputPath)
		hashInput := fmt.Sprintf("%s_%s_%s_%s_%s", idempotencyKey, req.Language, req.TargetLanguage, req.MacAddress, audioHash)
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
				utils.LogInfo("Pipeline: Duplicate request detected for IdempotencyKey %s. Returning existing TaskID %s", idempotencyKey, existingTaskID)
				return existingTaskID, nil
			}
			utils.LogInfo("Pipeline: Existing task %s found for IdempotencyKey %s but it is in 'failed' state or could not be loaded. Starting new task.", existingTaskID, idempotencyKey)
		}
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

	go u.runPipelineAsync(ctx, taskID, inputPath, req)

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

	// Stage 1: Transcription
	startTime := time.Now()
	status.Stages["transcription"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
	u.saveStatus(taskID, *status)

	refine := true
	if req.Refine != nil && !*req.Refine {
		refine = false
	}

	transResult, err := u.transcribeUC.TranscribeAudioSync(ctx, inputPath, req.Language, req.Diarize, refine)
	if err != nil {
		u.failStage(taskID, "transcription", err)
		return
	}
	status.Stages["transcription"] = pipelineDtos.PipelineStageStatus{
		Status:          "completed",
		Result:          transResult.Transcription,
		DurationSeconds: time.Since(startTime).Seconds(),
	}
	u.saveStatus(taskID, *status)

	// Stage 2: Refinement
	refinedText := transResult.Transcription
	if status.Stages["refinement"].Status != "skipped" {
		startTime = time.Now()
		status.Stages["refinement"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)

		// Refine is already called inside TranscribeAudioSync and returned as transResult.RefinedText
		// But if we want to be explicit or if we separated them:
		refinedText = transResult.RefinedText
		status.Stages["refinement"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          refinedText,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		u.saveStatus(taskID, *status)
	}

	// Stage 3: Translation
	finalText := refinedText
	if status.Stages["translation"].Status != "skipped" {
		startTime = time.Now()
		status.Stages["translation"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)

		transText, err := u.translateUC.TranslateTextSync(ctx, refinedText, req.TargetLanguage)
		if err != nil {
			u.failStage(taskID, "translation", err)
			return
		}
		finalText = transText
		status.Stages["translation"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          finalText,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		u.saveStatus(taskID, *status)
	}

	// Stage 4: Summary
	if status.Stages["summary"].Status != "skipped" {
		startTime = time.Now()
		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{Status: "processing", StartedAt: startTime.Format(time.RFC3339)}
		u.saveStatus(taskID, *status)

		participantsStr := strings.Join(req.Participants, ", ")
		summResult, err := u.summaryUC.SummarizeTextSync(ctx, finalText, req.TargetLanguage, req.Context, req.Style, req.Date, req.Location, participantsStr, req.MacAddress)
		if err != nil {
			u.failStage(taskID, "summary", err)
			return
		}
		status.Stages["summary"] = pipelineDtos.PipelineStageStatus{
			Status:          "completed",
			Result:          summResult,
			DurationSeconds: time.Since(startTime).Seconds(),
		}
		u.saveStatus(taskID, *status)
	}

	// Finalize
	status.OverallStatus = "completed"
	start, _ := time.Parse(time.RFC3339, status.StartedAt)
	status.DurationSeconds = time.Since(start).Seconds()
	u.saveStatus(taskID, *status)
}

func (u *pipelineUseCase) saveStatus(taskID string, status pipelineDtos.PipelineStatusDTO) {
	u.store.Set(taskID, &status)
	_ = u.cache.Set(taskID, status)
}

func (u *pipelineUseCase) failStage(taskID string, stageName string, err error) {
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
}
