package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
)

// Translate translates the given text to English using the LLM.
func (u *RAGUsecase) Translate(text string) (string, error) {
	prompt := fmt.Sprintf(`You are a professional translator and editor. 
Translate the following transcribed text to clear, grammatically correct English.
If the text is already in English, fix any grammatical errors and improve the clarity.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
English:`, text)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	translated, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Translate: original='%s', translated='%s', model='%s'", text, translated, model)
	return strings.TrimSpace(translated), nil
}
