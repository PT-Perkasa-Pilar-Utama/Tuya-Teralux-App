package usecases

import (
	"fmt"
	"os"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
	"time"

	"github.com/google/uuid"
)

type geminiServiceClient interface {
	HealthCheck() bool
	Transcribe(audioPath string, language string) (*dtos.WhisperResult, error)
}

type TranscribeGeminiModelUseCase interface {
	TranscribeAsync(filePath, fileName, language string, trigger ...string) (string, error)
}

type transcribeGeminiModelUseCase struct {
	service geminiServiceClient
	store   *tasks.StatusStore[dtos.AsyncTranscriptionStatusDTO]
	cache   *tasks.BadgerTaskCache
	config  *utils.Config
}

func NewTranscribeGeminiModelUseCase(
	service geminiServiceClient,
	store *tasks.StatusStore[dtos.AsyncTranscriptionStatusDTO],
	cache *tasks.BadgerTaskCache,
	cfg *utils.Config,
) TranscribeGeminiModelUseCase {
	return &transcribeGeminiModelUseCase{
		service: service,
		store:   store,
		cache:   cache,
		config:  cfg,
	}
}

func (u *transcribeGeminiModelUseCase) TranscribeAsync(filePath, fileName, language string, trigger ...string) (string, error) {
	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	taskID := uuid.New().String()
	triggerURL := ""
	if len(trigger) > 0 {
		triggerURL = trigger[0]
	}

	status := &dtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		Trigger:   triggerURL,
		StartedAt: time.Now().Format(time.RFC3339),
	}

	u.store.Set(taskID, status)
	_ = u.cache.Set(taskID, status)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.LogError("Gemini Task %s: Panic recovered: %v", taskID, r)
				u.updateStatus(taskID, "failed", nil, fmt.Errorf("internal panic: %v", r))
			}
		}()
		// Step 1: Health Check
		if !u.service.HealthCheck() {
			utils.LogError("Gemini Task %s: Service health check failed", taskID)
			u.updateStatus(taskID, "failed", nil, fmt.Errorf("Gemini service health check failed"))
			return
		}

		// Step 2: Transcribe
		result, err := u.service.Transcribe(filePath, language)
		if err != nil {
			utils.LogError("Gemini Task %s: Transcription failed: %v", taskID, err)
			u.updateStatus(taskID, "failed", nil, err)
			return
		}

		finalResult := &dtos.AsyncTranscriptionResultDTO{
			Transcription:    result.Transcription,
			DetectedLanguage: result.DetectedLanguage,
		}
		u.updateStatus(taskID, "completed", finalResult, nil)
	}()

	return taskID, nil
}

func (u *transcribeGeminiModelUseCase) updateStatus(taskID, statusStr string, result *dtos.AsyncTranscriptionResultDTO, err error) {
	// Try to preserve StartedAt
	var existing dtos.AsyncTranscriptionStatusDTO
	_, _, _ = u.cache.GetWithTTL(taskID, &existing)

	status := &dtos.AsyncTranscriptionStatusDTO{
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

	if statusStr == "completed" || statusStr == "failed" {
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			status.DurationSeconds = time.Since(startTime).Seconds()
		}
	}

	u.store.Set(taskID, status)
	_ = u.cache.SetPreserveTTL(taskID, status)
}
