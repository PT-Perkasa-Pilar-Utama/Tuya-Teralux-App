package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/skills"
	"sensio/domain/models/rag/skills/orchestrator"
	"strings"
)

type ChatUseCase interface {
	Chat(ctx context.Context, uid, terminalID, prompt, language string) (*dtos.RAGChatResponseDTO, error)
}

type ChatUseCaseImpl struct {
	llm          skills.LLMClient
	fallbackLLM  skills.LLMClient
	config       *utils.Config
	badger       *infrastructure.BadgerService
	vector       *infrastructure.VectorService
	orchestrator *orchestrator.Router
}

func NewChatUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, badger *infrastructure.BadgerService, vector *infrastructure.VectorService, orchestrator *orchestrator.Router) ChatUseCase {
	return &ChatUseCaseImpl{
		llm:          llm,
		fallbackLLM:  fallbackLLM,
		config:       cfg,
		badger:       badger,
		vector:       vector,
		orchestrator: orchestrator,
	}
}

func (u *ChatUseCaseImpl) Chat(ctx context.Context, uid, terminalID, prompt, language string) (*dtos.RAGChatResponseDTO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	// 1. Get History
	historyKey := fmt.Sprintf("chat_history:%s", terminalID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
		}
	}

	// 2. Prepare Skill Context
	skillCtx := &skills.SkillContext{
		Ctx:        ctx,
		UID:        uid,
		TerminalID: terminalID,
		Prompt:     prompt,
		Language:   language,
		History:    history,
		LLM:        u.llm,
		Config:     u.config,
		Vector:     u.vector,
		Badger:     u.badger,
	}

	// 3. Route and Execute via Orchestrator
	result, err := u.orchestrator.RouteAndExecute(skillCtx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("Chat: Primary LLM failed, falling back to local model: %v", err)
		skillCtx.LLM = u.fallbackLLM
		result, err = u.orchestrator.RouteAndExecute(skillCtx)
	}

	if err != nil {
		return nil, fmt.Errorf("orchestrator execution failed: %w", err)
	}

	// 4. Update History (skip if blocked)
	if u.badger != nil && !result.IsBlocked {
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
				Prompt:     prompt,
				TerminalID: terminalID,
			},
		}
	}

	return &dtos.RAGChatResponseDTO{
		Response:       result.Message,
		IsControl:      result.IsControl,
		IsBlocked:      result.IsBlocked,
		Redirect:       redirect,
		HTTPStatusCode: result.HTTPStatusCode,
	}, nil
}
