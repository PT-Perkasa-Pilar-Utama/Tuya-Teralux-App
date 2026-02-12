package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"

	"github.com/google/uuid"
)

// TranslateSync (private internal for use by Async)
func (u *RAGUsecase) translateInternal(text, targetLang string) (string, error) {
	langName := "English"
	if strings.ToLower(targetLang) == "id" {
		langName = "Indonesian"
	}

	prompt := fmt.Sprintf(`You are a professional translator and editor. 
Translate the following transcribed text to clear, grammatically correct %s.
If the text is already in %s, fix any grammatical errors and improve the clarity.
Only return the final polished text without any explanation, quotes, or additional commentary.

Text: "%s"
%s:`, langName, langName, text, langName)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	translated, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Translate: original='%s', translated='%s', model='%s', target='%s'", text, translated, model, langName)
	return strings.TrimSpace(translated), nil
}

func (u *RAGUsecase) Translate(text, targetLang string) (string, error) {
	return u.translateInternal(text, targetLang)
}

func (u *RAGUsecase) TranslateAsync(text, targetLang string) (string, error) {
	taskID := uuid.New().String()
	u.mu.Lock()
	u.taskStatus[taskID] = &dtos.RAGStatusDTO{Status: "pending"}
	u.mu.Unlock()

	if u.badger != nil {
		b, _ := json.Marshal(u.taskStatus[taskID])
		_ = u.badger.Set("rag:task:"+taskID, b)
	}

	go func() {
		translated, err := u.translateInternal(text, targetLang)
		u.mu.Lock()
		if err != nil {
			u.taskStatus[taskID] = &dtos.RAGStatusDTO{Status: "error", Result: err.Error()}
		} else {
			u.taskStatus[taskID] = &dtos.RAGStatusDTO{Status: "done", Result: translated}
		}
		status := u.taskStatus[taskID]
		u.mu.Unlock()

		if u.badger != nil {
			b, _ := json.Marshal(status)
			_ = u.badger.SetPreserveTTL("rag:task:"+taskID, b)
		}
	}()

	return taskID, nil
}
