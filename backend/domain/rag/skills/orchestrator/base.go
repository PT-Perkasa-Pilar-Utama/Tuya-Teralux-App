package orchestrator

import (
	"sensio/domain/rag/skills"
	"strings"
)

// BaseOrchestrator provides default implementation for placeholder replacement and LLM execution.
type BaseOrchestrator struct{}

func NewBaseOrchestrator() *BaseOrchestrator {
	return &BaseOrchestrator{}
}

func (b *BaseOrchestrator) Execute(ctx *skills.SkillContext, prompt string) (*skills.SkillResult, error) {
	// 1. Identify model to use
	model := "high"
	// Get skill name from context if possible, but orchestrator shouldn't strictly depend on name
	// However, for certain logic we might need hints.
	// For now let's just use generic logic.

	finalPrompt := prompt
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{prompt}}", ctx.Prompt)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{history}}", strings.Join(ctx.History, "\n"))

	// Special handling for Translation placeholders if present
	if strings.Contains(finalPrompt, "{{target_lang}}") {
		targetLang := "English"
		if ctx.Language != "" {
			switch {
			case strings.EqualFold(ctx.Language, "id") || strings.EqualFold(ctx.Language, "indonesian"):
				targetLang = "Indonesian"
			case strings.EqualFold(ctx.Language, "en") || strings.EqualFold(ctx.Language, "english"):
				targetLang = "English"
			default:
				targetLang = ctx.Language
			}
		} else if strings.Contains(strings.ToLower(ctx.Prompt), "indonesia") || strings.Contains(strings.ToLower(ctx.Prompt), "indo") {
			targetLang = "Indonesian"
		}
		finalPrompt = strings.ReplaceAll(finalPrompt, "{{target_lang}}", targetLang)
		model = "low"
	}

	// 2. Call LLM
	res, err := ctx.LLM.CallModel(finalPrompt, model)
	if err != nil {
		return nil, err
	}

	return &skills.SkillResult{
		Message:        strings.TrimSpace(res),
		HTTPStatusCode: 200,
	}, nil
}
