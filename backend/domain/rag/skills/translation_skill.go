package skills

import (
	"fmt"
	"strings"
)

// TranslationSkill handles requests to translate text between languages.
type TranslationSkill struct{}

func (s *TranslationSkill) Name() string {
	return "Translation"
}

func (s *TranslationSkill) Description() string {
	return "Handles requests to translate text between Indonesian and English, or to improve text clarity and grammar."
}

func (s *TranslationSkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	// Detect target language from context or prompt
	targetLang := "English"
	if ctx.Language != "" {
		switch {
		case strings.EqualFold(ctx.Language, "id") || strings.EqualFold(ctx.Language, "indonesian"):
			targetLang = "Indonesian"
		case strings.EqualFold(ctx.Language, "en") || strings.EqualFold(ctx.Language, "english"):
			targetLang = "English"
		default:
			targetLang = ctx.Language // Trust the explicit language code/name
		}
	} else if strings.Contains(strings.ToLower(ctx.Prompt), "indonesia") || strings.Contains(strings.ToLower(ctx.Prompt), "indo") {
		targetLang = "Indonesian"
	}

	prompt := fmt.Sprintf(`You are a professional translator and editor. 
Translate the following transcribed text to clear, grammatically correct %s.
If the text is already in %s, fix any grammatical errors and improve the clarity.
CRITICAL: Do not mention "Tuya" or "Tuya API" in your response. Use generic terms like "Smart Home System" or "Gateway" if needed.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
%s:`, targetLang, targetLang, ctx.Prompt, targetLang)

	model := "low"

	res, err := ctx.LLM.CallModel(prompt, model)
	if err != nil {
		return nil, err
	}

	return &SkillResult{
		Message:        strings.TrimSpace(res),
		HTTPStatusCode: 200,
	}, nil
}
