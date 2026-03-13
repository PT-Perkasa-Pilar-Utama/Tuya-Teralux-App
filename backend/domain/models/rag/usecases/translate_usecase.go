package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sensio/domain/common/providers"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/skills"
	"time"

	"github.com/google/uuid"
)

type TranslateUseCase interface {
	TranslateText(text, targetLang string, args ...string) (string, error)
	TranslateTextWithTrigger(text, targetLang string, trigger string, args ...string) (string, error)
	TranslateTextSync(ctx context.Context, text, targetLang string, args ...string) (string, error)
}

type translateUseCase struct {
	llm              skills.LLMClient
	fallbackLLM      skills.LLMClient
	config           *utils.Config
	cache            *tasks.BadgerTaskCache
	store            *tasks.StatusStore[dtos.RAGStatusDTO]
	mqttSvc          mqttPublisher
	skill            skills.Skill
	providerResolver providers.ProviderResolver
}

func NewTranslateUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, cache *tasks.BadgerTaskCache, store *tasks.StatusStore[dtos.RAGStatusDTO], mqttSvc mqttPublisher, skill skills.Skill, providerResolver providers.ProviderResolver) TranslateUseCase {
	return &translateUseCase{
		llm:              llm,
		fallbackLLM:      fallbackLLM,
		config:           cfg,
		cache:            cache,
		store:            store,
		mqttSvc:          mqttSvc,
		skill:            skill,
		providerResolver: providerResolver,
	}
}

// translateInternal (private internal for use by Execute)
// Optional: pass macAddress as third argument for terminal-specific provider resolution
func (u *translateUseCase) translateInternal(ctx context.Context, text, targetLang string, args ...string) (string, error) {
	if u.skill == nil {
		return "", fmt.Errorf("translation skill not configured")
	}

	// Resolve provider based on macAddress if provided
	llmClient := u.llm
	fallbackClient := u.fallbackLLM
	if u.providerResolver != nil && len(args) > 0 && args[0] != "" {
		macAddress := args[0]
		resolved, err := u.providerResolver.ResolveByMacAddress(macAddress)
		if err != nil {
			utils.LogWarn("TranslateUseCase: Provider resolution failed for MAC %s: %v, using default", macAddress, err)
		} else {
			llmClient = resolved.LLM
			fallbackClient = resolved.FallbackLLM
			utils.LogInfo("TranslateUseCase: Using terminal-specific provider '%s' for MAC %s", resolved.ProviderName, macAddress)
		}
	}

	skillCtx := &skills.SkillContext{
		Ctx:      ctx,
		Prompt:   text,
		Language: targetLang,
		LLM:      llmClient,
		Config:   u.config,
	}

	res, err := u.skill.Execute(skillCtx)
	if err != nil && fallbackClient != nil {
		utils.LogWarn("Translate: Primary LLM failed, falling back to local model: %v", err)
		skillCtx.LLM = fallbackClient
		res, err = u.skill.Execute(skillCtx)
	}

	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Translate: original='%s', translated='%s', target='%s'", text, res.Message, targetLang)
	return res.Message, nil
}

func (u *translateUseCase) TranslateTextSync(ctx context.Context, text, targetLang string, args ...string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return u.translateInternal(ctx, text, targetLang, args...)
}

func (u *translateUseCase) TranslateText(text, targetLang string, args ...string) (string, error) {
	return u.TranslateTextWithTrigger(text, targetLang, "", args...)
}

// args is passed as [0] macAddress, [1] idempotencyKey
func (u *translateUseCase) TranslateTextWithTrigger(text, targetLang string, trigger string, args ...string) (string, error) {
	mac := ""
	idempotencyKey := ""
	if len(args) > 0 {
		mac = args[0]
	}
	if len(args) > 1 {
		idempotencyKey = args[1]
	}

	// 1. Idempotency Check
	var idempotencyHash string
	if idempotencyKey != "" {
		textHash := utils.HashString(text)
		hashInput := fmt.Sprintf("%s_%s_%s_%s", idempotencyKey, targetLang, mac, textHash)
		idempotencyHash = "idemp_translate_" + utils.HashString(hashInput)

		var existingTaskID string
		if _, exists, _ := u.cache.GetWithTTL(idempotencyHash, &existingTaskID); exists && existingTaskID != "" {
			// Check task state - fallback to cache if store is empty
			status, ok := u.store.Get(existingTaskID)
			if !ok || status == nil {
				var cachedStatus dtos.RAGStatusDTO
				if _, cachedExists, _ := u.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
					status = &cachedStatus
					ok = true
					// Prime the store for future calls
					u.store.Set(existingTaskID, status)
				}
			}

			if ok && status != nil && status.Status != "failed" {
				utils.LogInfo("Translate Task: Duplicate request detected for IdempotencyKey %s. Returning existing TaskID %s (Status: %s)", idempotencyKey, existingTaskID, status.Status)
				return existingTaskID, nil
			}
			utils.LogInfo("Translate Task: Found existing task %s for key %s but it failed or could not be loaded. Proceeding with new task.", existingTaskID, idempotencyKey)
		}
	}

	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{
		Status:     "pending",
		Trigger:    trigger,
		MacAddress: mac,
		StartedAt:  time.Now().Format(time.RFC3339),
	}
	u.store.Set(taskID, status)
	_ = u.cache.Set(taskID, status)

	if idempotencyHash != "" {
		// Store mapping
		_ = u.cache.Set(idempotencyHash, taskID)
	}

	go func() {
		translated, err := u.translateInternal(context.Background(), text, targetLang)
		if err != nil {
			utils.LogError("RAG Translate Task %s: Failed with error: %v", taskID, err)
			u.updateStatus(taskID, "failed", err, "")
		} else {
			utils.LogInfo("RAG Translate Task %s: Completed successfully", taskID)
			u.updateStatus(taskID, "completed", nil, translated)
		}
	}()

	return taskID, nil
}

func (u *translateUseCase) updateStatus(taskID string, statusStr string, err error, result string) {
	// Try to get existing status to preserve StartedAt and MacAddress
	var existing dtos.RAGStatusDTO
	_, _, _ = u.cache.GetWithTTL(taskID, &existing)

	status := &dtos.RAGStatusDTO{
		Status:     statusStr,
		StartedAt:  existing.StartedAt,
		Trigger:    existing.Trigger,
		MacAddress: existing.MacAddress,
		ExpiresAt:  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if err != nil {
		status.Error = err.Error()
		status.Result = err.Error()
		status.HTTPStatusCode = utils.GetErrorStatusCode(err)
	}

	if result != "" {
		status.Result = result
		status.HTTPStatusCode = 200
	}

	if statusStr == "completed" || statusStr == "failed" {
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			status.DurationSeconds = time.Since(startTime).Seconds()
		}

		// Send MQTT "stop" signal if MacAddress is available
		if status.MacAddress != "" && u.mqttSvc != nil {
			taskTopic := fmt.Sprintf("users/%s/%s/task", status.MacAddress, u.config.ApplicationEnvironment)
			msg := map[string]string{
				"event": "stop",
				"task":  "RAG",
			}
			payload, _ := json.Marshal(msg)
			if err := u.mqttSvc.Publish(taskTopic, 0, false, payload); err != nil {
				utils.LogError("RAG Translate Task %s: Failed to publish stop signal to MQTT: %v", taskID, err)
			} else {
				utils.LogInfo("RAG Translate Task %s: Published stop signal to %s", taskID, taskTopic)
			}
		}
	}

	u.store.Set(taskID, status)
	_ = u.cache.SetPreserveTTL(taskID, status)
}
