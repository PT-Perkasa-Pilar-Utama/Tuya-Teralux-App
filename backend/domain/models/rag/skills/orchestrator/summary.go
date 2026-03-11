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

	finalPrompt := prompt
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{prompt}}", ctx.Prompt)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{history}}", strings.Join(ctx.History, "\n"))
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{audience_guidance}}", promptConfig.AudienceGuidance())
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{risk_scoring_guidance}}", promptConfig.RiskScoringGuidance())
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{assertiveness_phrasing}}", promptConfig.AssertivenessPhrasing())
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{language}}", targetLangName)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{context}}", ctx.Context)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{style}}", ctx.Style)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{date}}", ctx.Date)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{location}}", ctx.Location)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{participants}}", ctx.Participants)

	res, err := ctx.LLM.CallModel(ctx.Ctx, finalPrompt, "high")
	if err != nil {
		return nil, err
	}

	return &skills.SkillResult{
		Message:        strings.TrimSpace(res),
		HTTPStatusCode: 200,
	}, nil
}
