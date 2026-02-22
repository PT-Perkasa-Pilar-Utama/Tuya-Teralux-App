package usecases

import (
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/skills"
)

type RefineUseCase interface {
	RefineText(text string, lang string) (string, error)
}

type refineUseCase struct {
	llm         skills.LLMClient
	fallbackLLM skills.LLMClient
	config      *utils.Config
}

func NewRefineUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config) RefineUseCase {
	return &refineUseCase{
		llm:         llm,
		fallbackLLM: fallbackLLM,
		config:      cfg,
	}
}

// RefineText improves the grammar and clarity of the transcribed text based on the detected language.
func (u *refineUseCase) RefineText(text string, lang string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	skill := &skills.RefineSkill{}
	ctx := &skills.SkillContext{
		Prompt:   text,
		Language: lang,
		LLM:      u.llm,
		Config:   u.config,
	}

	res, err := skill.Execute(ctx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("Refine: Primary LLM failed, falling back to local model: %v", err)
		ctx.LLM = u.fallbackLLM
		res, err = skill.Execute(ctx)
	}

	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Refine: lang='%s', original='%s', refined='%s'", lang, text, res.Message)
	return res.Message, nil
}
