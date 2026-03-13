package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/providers"
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
	llm              skills.LLMClient
	fallbackLLM      skills.LLMClient
	config           *utils.Config
	badger           *infrastructure.BadgerService
	vector           *infrastructure.VectorService
	orchestrator     *orchestrator.Router
	providerResolver providers.ProviderResolver
}

func NewChatUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, badger *infrastructure.BadgerService, vector *infrastructure.VectorService, orchestrator *orchestrator.Router, providerResolver providers.ProviderResolver) ChatUseCase {
	return &ChatUseCaseImpl{
		llm:              llm,
		fallbackLLM:      fallbackLLM,
		config:           cfg,
		badger:           badger,
		vector:           vector,
		orchestrator:     orchestrator,
		providerResolver: providerResolver,
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

	// 2. Resolve provider based on terminal ID
	llmClient := u.llm
	fallbackClient := u.fallbackLLM
	if u.providerResolver != nil && terminalID != "" {
		resolved, err := u.providerResolver.ResolveByTerminalID(terminalID)
		if err != nil {
			utils.LogWarn("ChatUseCase: Provider resolution failed for terminal %s: %v, using default", terminalID, err)
		} else {
			llmClient = resolved.LLM
			fallbackClient = resolved.FallbackLLM
			utils.LogInfo("ChatUseCase: Using terminal-specific provider '%s' for terminal %s", resolved.ProviderName, terminalID)
		}
	}

	// 3. Prepare Skill Context
	skillCtx := &skills.SkillContext{
		Ctx:        ctx,
		UID:        uid,
		TerminalID: terminalID,
		Prompt:     prompt,
		Language:   language,
		History:    history,
		LLM:        llmClient,
		Config:     u.config,
		Vector:     u.vector,
		Badger:     u.badger,
	}

	// 4. Route and Execute via Orchestrator
	result, err := u.orchestrator.RouteAndExecute(skillCtx)
	if err != nil && fallbackClient != nil {
		utils.LogWarn("Chat: Primary LLM failed, falling back to local model: %v", err)
		skillCtx.LLM = fallbackClient
		result, err = u.orchestrator.RouteAndExecute(skillCtx)
	}

	// If orchestrator still fails, fallback to ServiceIssue skill for graceful error handling
	if err != nil {
		utils.LogWarn("Chat: Orchestrator failed after fallback, using ServiceIssue skill for graceful response: %v", err)
		serviceIssueSkill, hasServiceIssue := u.orchestrator.GetSkillRegistry().Get("ServiceIssue")
		if hasServiceIssue {
			serviceIssueCtx := &skills.SkillContext{
				Ctx:        ctx,
				UID:        uid,
				TerminalID: terminalID,
				Prompt:     "Service unavailable",
				Language:   language,
				History:    history,
				LLM:        fallbackClient,
				Config:     u.config,
				Vector:     u.vector,
				Badger:     u.badger,
			}
			serviceIssueResult, serviceIssueExecErr := serviceIssueSkill.Execute(serviceIssueCtx)
			if serviceIssueExecErr == nil {
				result = serviceIssueResult
			} else {
				// Last resort: return a simple service-issue response
				result = &skills.SkillResult{
					Message:   "Maaf, koneksi atau layanan AI sedang bermasalah. Coba lagi sebentar ya.",
					IsBlocked: false,
				}
			}
		} else {
			// ServiceIssue skill not available, return simple response
			result = &skills.SkillResult{
				Message:   "Maaf, koneksi atau layanan AI sedang bermasalah. Coba lagi sebentar ya.",
				IsBlocked: false,
			}
		}
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
