package skills

import (
	"fmt"
	"sensio/domain/rag/services"
	"strings"
)

// SummarySkill handles requests to summarize text or meetings.
type SummarySkill struct{}

func (s *SummarySkill) Name() string {
	return "Summary"
}

func (s *SummarySkill) Description() string {
	return "Summarizes text or meeting transcripts into structured reports."
}

func (s *SummarySkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	text := ctx.Prompt
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text to summarize is empty")
	}

	// Default to Indonesian if not specified
	targetLangName := "Indonesian"
	if strings.EqualFold(ctx.Language, "en") {
		targetLangName = "English"
	}

	// Metadata for MoM
	date := ctx.Date
	location := ctx.Location
	participants := ctx.Participants
	meetingContext := ctx.Context
	if meetingContext == "" {
		meetingContext = "meeting"
	}
	style := ctx.Style
	if style == "" {
		style = "executive"
	}

	// Build structured prompt using the configuration service
	promptConfig := &services.PromptConfig{
		Assertiveness: 8,          // Strategic assertiveness
		Audience:      "mixed",    // C-level + VP/Director level
		RiskScale:     "granular", // 1-10 scoring
		Context:       meetingContext,
		Style:         style,
		Language:      targetLangName,
		Date:          date,
		Location:      location,
		Participants:  participants,
	}

	prompt := promptConfig.BuildPrompt(text)
	model := "high" // Summarization needs high intelligence

	summary, err := ctx.LLM.CallModel(prompt, model)
	if err != nil {
		return nil, fmt.Errorf("summary skill failed: %w", err)
	}

	return &SkillResult{
		Message:        strings.TrimSpace(summary),
		HTTPStatusCode: 200,
	}, nil
}
