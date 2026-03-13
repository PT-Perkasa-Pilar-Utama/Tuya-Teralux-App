package usecases

import (
	"context"
	"fmt"
	"sensio/domain/common/providers"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	"strings"
	"time"
)

type RefineUseCase interface {
	RefineText(ctx context.Context, text string, lang string, args ...string) (string, error)
}

type refineUseCase struct {
	llm              skills.LLMClient
	fallbackLLM      skills.LLMClient
	config           *utils.Config
	skill            skills.Skill
	providerResolver providers.ProviderResolver
}

func NewRefineUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, skill skills.Skill, providerResolver providers.ProviderResolver) RefineUseCase {
	return &refineUseCase{
		llm:              llm,
		fallbackLLM:      fallbackLLM,
		config:           cfg,
		skill:            skill,
		providerResolver: providerResolver,
	}
}

// RefineText improves the grammar and clarity of the transcribed text based on the detected language.
// Optional: pass macAddress as third argument for terminal-specific provider resolution
func (u *refineUseCase) RefineText(ctx context.Context, text string, lang string, args ...string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	textChars := len(text)
	startTime := time.Now()
	utils.LogDebug("Refine: starting (lang=%s chars=%d)", lang, textChars)

	if u.skill == nil {
		return "", fmt.Errorf("refine skill not configured")
	}

	// Resolve provider based on macAddress if provided
	llmClient := u.llm
	fallbackClient := u.fallbackLLM
	if u.providerResolver != nil && len(args) > 0 && args[0] != "" {
		macAddress := args[0]
		resolved, err := u.providerResolver.ResolveByMacAddress(macAddress)
		if err != nil {
			utils.LogWarn("RefineUseCase: Provider resolution failed for MAC %s: %v, using default", macAddress, err)
		} else {
			llmClient = resolved.LLM
			fallbackClient = resolved.FallbackLLM
			utils.LogInfo("RefineUseCase: Using terminal-specific provider '%s' for MAC %s", resolved.ProviderName, macAddress)
		}
	}

	skillCtx := &skills.SkillContext{
		Ctx:      ctx,
		Prompt:   text,
		Language: lang,
		LLM:      llmClient,
		Config:   u.config,
	}

	res, err := u.skill.Execute(skillCtx)
	if err != nil && fallbackClient != nil {
		utils.LogWarn("Refine: Primary LLM failed, falling back to local model: %v", err)
		skillCtx.LLM = fallbackClient
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
