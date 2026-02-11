package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
)

// Refine improves the grammar and clarity of the transcribed text based on the detected language.
func (u *RAGUsecase) Refine(text string, lang string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	var prompt string

	if strings.ToLower(lang) == "id" {
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

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	refined, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Refine: lang='%s', original='%s', refined='%s', model='%s'", lang, text, refined, model)
	return strings.TrimSpace(refined), nil
}
