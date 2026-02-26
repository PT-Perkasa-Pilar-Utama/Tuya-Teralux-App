package skills

import (
	"fmt"
	"strings"
)

// RefineSkill handles requests to improve grammar and clarity of text.
type RefineSkill struct{}

func (s *RefineSkill) Name() string {
	return "Refine"
}

func (s *RefineSkill) Description() string {
	return "Improves the grammar, spelling, and clarity of text in Indonesian or English."
}

func (s *RefineSkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	text := ctx.Prompt
	if strings.TrimSpace(text) == "" {
		return &SkillResult{Message: ""}, nil
	}

	prompt := fmt.Sprintf(`You are a professional editor. 
Fix any grammatical errors, spelling mistakes, and improve the clarity of the following transcription.
Maintain the original language of the text (e.g., if it is in Indonesian, keep it in Indonesian but fix the grammar).
If the text is already clear and correct, return it as is.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
Refined Text:`, text)

	model := "low" // Low model is sufficient for grammar fixing

	refined, err := ctx.LLM.CallModel(prompt, model)
	if err != nil {
		return nil, fmt.Errorf("refine skill failed: %w", err)
	}

	return &SkillResult{
		Message:        strings.TrimSpace(refined),
		HTTPStatusCode: 200,
	}, nil
}
