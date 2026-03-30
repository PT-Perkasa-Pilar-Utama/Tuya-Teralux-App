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

	// Use centralized health-aware fallback chain with terminal preference if macAddress provided
	var result string
	var err error

	if len(args) > 0 && args[0] != "" {
		// Use terminal-specific provider preference
		macAddress := args[0]
		err = u.providerResolver.ExecuteWithFallbackByMac(macAddress, func(resolvedSet *providers.ResolvedProviderSet) error {
			skillCtx := &skills.SkillContext{
				Ctx:      ctx,
				Prompt:   text,
				Language: lang,
				LLM:      resolvedSet.LLM,
				Config:   u.config,
			}
			res, execErr := u.skill.Execute(skillCtx)
			if execErr == nil {
				result = res.Message
			}
			return execErr
		})
	} else {
		// Use standard health-aware fallback
		err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *providers.ResolvedProviderSet) error {
			skillCtx := &skills.SkillContext{
				Ctx:      ctx,
				Prompt:   text,
				Language: lang,
				LLM:      resolvedSet.LLM,
				Config:   u.config,
			}
			res, execErr := u.skill.Execute(skillCtx)
			if execErr == nil {
				result = res.Message
			}
			return execErr
		})
	}

	if err != nil {
		utils.LogWarn("Refine: failed (lang=%s chars=%d duration=%s) err=%v", lang, textChars, time.Since(startTime), err)
		return "", err
	}

	utils.LogDebug("Refine: completed (lang=%s chars=%d duration=%s output_chars=%d)", lang, textChars, time.Since(startTime), len(result))
	utils.LogDebug("RAG Refine: lang='%s', original='%s', refined='%s'", lang, text, result)
	return result, nil
}
