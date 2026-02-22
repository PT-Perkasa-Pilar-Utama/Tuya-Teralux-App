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

	var prompt string
	if strings.EqualFold(ctx.Language, "id") {
		// Indonesian KBBI / Formal Fix
		prompt = fmt.Sprintf(`You are a professional Indonesian editor. 
Fix the grammar, spelling, and word choices of the following transcription to align with standard Indonesian (KBBI/PUEBI).
Ensure the tone is clear and professional while preserving the original meaning.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
Sesuai KBBI:`, text)
	} else {
		// English / Default Grammar Fix
		prompt = fmt.Sprintf(`You are a professional English editor. 
Fix any grammatical errors, spelling mistakes, and improve the clarity of the following transcription.
If the text is already clear, return it as is.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
Refined English:`, text)
	}

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
