package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/utilities"

	"github.com/google/uuid"
)

type TranslateUseCase interface {
	TranslateText(text, targetLang string) (string, error)
	TranslateTextSync(text, targetLang string) (string, error)
}

type translateUseCase struct {
	llm    utilities.LLMClient
	config *utils.Config
	cache  *tasks.BadgerTaskCache
	store  *tasks.StatusStore[dtos.RAGStatusDTO]
}

func NewTranslateUseCase(llm utilities.LLMClient, cfg *utils.Config, cache *tasks.BadgerTaskCache, store *tasks.StatusStore[dtos.RAGStatusDTO]) TranslateUseCase {
	return &translateUseCase{
		llm:    llm,
		config: cfg,
		cache:  cache,
		store:  store,
	}
}

// translateInternal (private internal for use by Execute)
func (u *translateUseCase) translateInternal(text, targetLang string) (string, error) {
	langName := "English"
	if strings.ToLower(targetLang) == "id" {
		langName = "Indonesian"
	}

	prompt := fmt.Sprintf(`You are a professional translator and editor. 
Translate the following transcribed text to clear, grammatically correct %s.
If the text is already in %s, fix any grammatical errors and improve the clarity.
CRITICAL: Do not mention "Tuya" or "Tuya API" in your response. Use generic terms like "Smart Home System" or "Gateway" if needed.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
%s:`, langName, langName, text, langName)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	translated, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Translate: original='%s', translated='%s', target='%s'", text, translated, langName)
	return strings.TrimSpace(translated), nil
}

func (u *translateUseCase) TranslateTextSync(text, targetLang string) (string, error) {
	return u.translateInternal(text, targetLang)
}

func (u *translateUseCase) TranslateText(text, targetLang string) (string, error) {
	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{Status: "pending"}
	u.store.Set(taskID, status)

	_ = u.cache.Set(taskID, status)

	go func() {
		translated, err := u.translateInternal(text, targetLang)
		var finalStatus *dtos.RAGStatusDTO
		if err != nil {
			utils.LogError("RAG Translate Task %s: Failed with error: %v", taskID, err)
			finalStatus = &dtos.RAGStatusDTO{Status: "failed", Result: err.Error()}
		} else {
			utils.LogInfo("RAG Translate Task %s: Completed successfully", taskID)
			finalStatus = &dtos.RAGStatusDTO{Status: "completed", Result: translated}
		}

		u.store.Set(taskID, finalStatus)
		_ = u.cache.SetPreserveTTL(taskID, finalStatus)
	}()

	return taskID, nil
}
