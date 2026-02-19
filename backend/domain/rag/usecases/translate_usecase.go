package usecases

import (
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/skills"
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
	skill := &skills.TranslationSkill{}
	ctx := &skills.SkillContext{
		Prompt:   text,
		Language: targetLang,
		LLM:      u.llm,
		Config:   u.config,
	}

	res, err := skill.Execute(ctx)
	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Translate: original='%s', translated='%s', target='%s'", text, res.Message, targetLang)
	return res.Message, nil
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
