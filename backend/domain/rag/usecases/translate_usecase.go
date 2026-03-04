package usecases

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/skills"
	"time"

	"github.com/google/uuid"
)

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

type TranslateUseCase interface {
	TranslateText(text, targetLang string, macAddress ...string) (string, error)
	TranslateTextWithTrigger(text, targetLang string, trigger string, macAddress ...string) (string, error)
	TranslateTextSync(text, targetLang string) (string, error)
}

type translateUseCase struct {
	llm         skills.LLMClient
	fallbackLLM skills.LLMClient
	config      *utils.Config
	cache       *tasks.BadgerTaskCache
	store       *tasks.StatusStore[dtos.RAGStatusDTO]
	mqttSvc     mqttPublisher
}

func NewTranslateUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, cache *tasks.BadgerTaskCache, store *tasks.StatusStore[dtos.RAGStatusDTO], mqttSvc mqttPublisher) TranslateUseCase {
	return &translateUseCase{
		llm:         llm,
		fallbackLLM: fallbackLLM,
		config:      cfg,
		cache:       cache,
		store:       store,
		mqttSvc:     mqttSvc,
	}
}

// translateInternal (private internal for use by Execute)
func (u *translateUseCase) translateInternal(text, targetLang string) (string, error) {
	skill := &skills.TranslationSkill{}
	ctx := &skills.SkillContext{
		Prompt:   text,
		Language: targetLang,
		LLM:      u.llm,
		Config:   u.config,
	}

	res, err := skill.Execute(ctx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("Translate: Primary LLM failed, falling back to local model: %v", err)
		ctx.LLM = u.fallbackLLM
		res, err = skill.Execute(ctx)
	}

	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Translate: original='%s', translated='%s', target='%s'", text, res.Message, targetLang)
	return res.Message, nil
}

func (u *translateUseCase) TranslateTextSync(text, targetLang string) (string, error) {
	return u.translateInternal(text, targetLang)
}

func (u *translateUseCase) TranslateText(text, targetLang string, macAddress ...string) (string, error) {
	return u.TranslateTextWithTrigger(text, targetLang, "", macAddress...)
}

func (u *translateUseCase) TranslateTextWithTrigger(text, targetLang string, trigger string, macAddress ...string) (string, error) {
	mac := ""
	if len(macAddress) > 0 {
		mac = macAddress[0]
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

	go func() {
		translated, err := u.translateInternal(text, targetLang)
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
			taskTopic := fmt.Sprintf("users/%s/task", status.MacAddress)
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
