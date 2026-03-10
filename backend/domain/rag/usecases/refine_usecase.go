package usecases

import (
	"context"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/rag/skills"
	"strings"
	"time"
)

type RefineUseCase interface {
	RefineText(ctx context.Context, text string, lang string) (string, error)
}

type refineUseCase struct {
	llm         skills.LLMClient
	fallbackLLM skills.LLMClient
	config      *utils.Config
	skill       skills.Skill
}

func NewRefineUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, skill skills.Skill) RefineUseCase {
	return &refineUseCase{
		llm:         llm,
		fallbackLLM: fallbackLLM,
		config:      cfg,
		skill:       skill,
	}
}

// RefineText improves the grammar and clarity of the transcribed text based on the detected language.
func (u *refineUseCase) RefineText(ctx context.Context, text string, lang string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	textChars := len(text)
	startTime := time.Now()
	utils.LogDebug("Refine: starting (lang=%s chars=%d)", lang, textChars)

	if u.skill == nil {
		return "", fmt.Errorf("refine skill not configured")
	}
	skillCtx := &skills.SkillContext{
		Ctx:      ctx,
		Prompt:   text,
		Language: lang,
		LLM:      u.llm,
		Config:   u.config,
	}

	res, err := u.skill.Execute(skillCtx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("Refine: Primary LLM failed, falling back to local model: %v", err)
		skillCtx.LLM = u.fallbackLLM
		res, err = u.skill.Execute(skillCtx)
	}

	if err != nil {
		utils.LogWarn("Refine: failed (lang=%s chars=%d duration=%s) err=%v", lang, textChars, time.Since(startTime), err)
		return "", err
	}

	utils.LogDebug("Refine: completed (lang=%s chars=%d duration=%s output_chars=%d)", lang, textChars, time.Since(startTime), len(res.Message))
	utils.LogDebug("RAG Refine: lang='%s', original='%s', refined='%s'", lang, text, res.Message)
	return res.Message, nil
}
