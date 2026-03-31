package orchestrator

import (
	"sensio/domain/models/rag/services"
	"sensio/domain/models/rag/skills"
	"strings"
)

type SummaryOrchestrator struct{}

func NewSummaryOrchestrator() *SummaryOrchestrator {
	return &SummaryOrchestrator{}
}

func (o *SummaryOrchestrator) Execute(ctx *skills.SkillContext, prompt string) (*skills.SkillResult, error) {
	targetLangName := "Indonesian"
	if strings.EqualFold(ctx.Language, "en") {
		targetLangName = "English"
	}

	promptConfig := &services.PromptConfig{
		Assertiveness: 8,
		Audience:      "mixed",
		RiskScale:     "granular",
		Context:       ctx.Context,
		Style:         ctx.Style,
		Language:      targetLangName,
		Date:          ctx.Date,
		Location:      ctx.Location,
		Participants:  ctx.Participants,
	}

	// Use style-aware prompt family instead of single generic prompt
	finalPrompt := promptConfig.GetStylePrompt(ctx.Prompt)

	res, err := ctx.LLM.CallModel(ctx.Ctx, finalPrompt, "high")
	if err != nil {
		return nil, err
	}

	return &skills.SkillResult{
		Message:        strings.TrimSpace(res),
		HTTPStatusCode: 200,
	}, nil
}
