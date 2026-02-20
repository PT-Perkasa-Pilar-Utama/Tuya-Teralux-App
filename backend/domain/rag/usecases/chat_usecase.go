package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/skills"
)

type ChatUseCase interface {
	Chat(uid, teraluxID, prompt, language string) (*dtos.RAGChatResponseDTO, error)
}

type ChatUseCaseImpl struct {
	llm          skills.LLMClient
	config       *utils.Config
	badger       *infrastructure.BadgerService
	vector       *infrastructure.VectorService
	orchestrator *skills.Orchestrator
}

func NewChatUseCase(llm skills.LLMClient, cfg *utils.Config, badger *infrastructure.BadgerService, vector *infrastructure.VectorService, orchestrator *skills.Orchestrator) ChatUseCase {
	return &ChatUseCaseImpl{
		llm:          llm,
		config:       cfg,
		badger:       badger,
		vector:       vector,
		orchestrator: orchestrator,
	}
}

func (u *ChatUseCaseImpl) Chat(uid, teraluxID, prompt, language string) (*dtos.RAGChatResponseDTO, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	// 1. Get History
	historyKey := fmt.Sprintf("chat_history:%s", teraluxID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
		}
	}

	// 2. Prepare Skill Context
	ctx := &skills.SkillContext{
		UID:       uid,
		TeraluxID: teraluxID,
		Prompt:    prompt,
		Language:  language,
		History:   history,
		LLM:       u.llm,
		Config:    u.config,
		Vector:    u.vector,
		Badger:    u.badger,
	}

	// 3. Route and Execute via Orchestrator
	result, err := u.orchestrator.RouteAndExecute(ctx)
	if err != nil {
		return nil, fmt.Errorf("orchestrator execution failed: %w", err)
	}

	// 4. Update History
	if u.badger != nil {
		history = append(history, "User: "+prompt, "Assistant: "+result.Message)
		if len(history) > 20 {
			history = history[len(history)-20:]
		}
		data, _ := json.Marshal(history)
		_ = u.badger.Set(historyKey, data)
	}

	// 5. Handle Redirect for Control skill
	var redirect *dtos.RedirectDTO
	if result.IsControl && result.Data != nil {
		redirect = &dtos.RedirectDTO{
			Endpoint: "/api/rag/control",
			Method:   "POST",
			Body: dtos.RAGControlRequestDTO{
				Prompt:    prompt,
				TeraluxID: teraluxID,
			},
		}
	}

	return &dtos.RAGChatResponseDTO{
		Response:       result.Message,
		IsControl:      result.IsControl,
		Redirect:       redirect,
		HTTPStatusCode: result.HTTPStatusCode,
	}, nil
}
