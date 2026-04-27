package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sensio/domain/common/interfaces"
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/skills"
	speechUsecases "sensio/domain/speech/usecases"
	"strings"
	"time"
)

type ControlUseCase interface {
	ProcessControl(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error)
}

type controlUseCase struct {
	llm              skills.LLMClient
	fallbackLLM      skills.LLMClient
	config           *utils.Config
	vector           *infrastructure.VectorService
	badger           *infrastructure.BadgerService
	tuyaExecutor     interfaces.DeviceControlExecutor
	tuyaAuth         interfaces.AuthUseCase
	skill            skills.Skill
	providerResolver speechUsecases.ProviderResolver
}

func NewControlUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, vector *infrastructure.VectorService, badger *infrastructure.BadgerService, tuyaExecutor interfaces.DeviceControlExecutor, tuyaAuth interfaces.AuthUseCase, skill skills.Skill, providerResolver speechUsecases.ProviderResolver) ControlUseCase {
	return &controlUseCase{
		llm:              llm,
		fallbackLLM:      fallbackLLM,
		config:           cfg,
		vector:           vector,
		badger:           badger,
		tuyaExecutor:     tuyaExecutor,
		tuyaAuth:         tuyaAuth,
		skill:            skill,
		providerResolver: providerResolver,
	}
}

func (u *controlUseCase) ProcessControl(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
	ucStart := time.Now()

	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	// Delegate control logic to ControlSkill
	if u.skill == nil {
		return nil, fmt.Errorf("control skill not configured")
	}

	// Provider resolution is handled by FallbackOrchestrator
	providerDuration := time.Millisecond * 0

	skillCtx := &skills.SkillContext{
		Ctx:        ctx,
		UID:        uid,
		TerminalID: terminalID,
		Prompt:     prompt,
		LLM:        u.llm, // Initialized with default, overridden by executeSkillWithFallback
		Config:     u.config,
		Vector:     u.vector,
		Badger:     u.badger,
	}

	// Preload history for the skill
	historyStart := time.Now()
	historyKey := fmt.Sprintf("chat_history:%s", terminalID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
			utils.LogDebug("ControlUseCase: History loaded | key=%s | duration_ms=%d | size=%d", historyKey, time.Since(historyStart).Milliseconds(), len(history))
		} else {
			utils.LogDebug("ControlUseCase: History not found | key=%s | duration_ms=%d", historyKey, time.Since(historyStart).Milliseconds())
		}
	}
	historyDuration := time.Since(historyStart)
	skillCtx.History = history

	// Execute skill (LLM call happens here)
	skillStart := time.Now()
	res, err := u.executeSkillWithFallback(ctx, skillCtx)
	skillDuration := time.Since(skillStart)

	utils.LogDebug("ControlUseCase: Skill.Execute completed | duration_ms=%d | err=%v", skillDuration.Milliseconds(), err)

	if err != nil {
		totalDuration := time.Since(ucStart)
		utils.LogError("ControlUseCase: ProcessControl failed | total_duration_ms=%d | error=%v", totalDuration.Milliseconds(), err)
		return nil, err
	}

	deviceID := ""
	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if id, ok := dataMap["device_id"].(string); ok {
			deviceID = id
		}
	}

	totalDuration := time.Since(ucStart)
	utils.LogInfo("ControlUseCase: ProcessControl completed | terminalID=%s | provider_duration_ms=%d | history_duration_ms=%d | skill_duration_ms=%d | total_duration_ms=%d | deviceID=%s",
		terminalID, providerDuration.Milliseconds(), historyDuration.Milliseconds(), skillDuration.Milliseconds(), totalDuration.Milliseconds(), deviceID)

	return &dtos.ControlResultDTO{
		Message:        res.Message,
		DeviceID:       deviceID,
		HTTPStatusCode: res.HTTPStatusCode,
	}, nil

}

// executeSkillWithFallback executes the skill with health-aware remote provider fallback
func (u *controlUseCase) executeSkillWithFallback(_ context.Context, skillCtx *skills.SkillContext) (*skills.SkillResult, error) {
	var result *skills.SkillResult
	var err error

	if skillCtx.TerminalID != "" {
		// Use terminal-specific provider preference
		err = u.providerResolver.ExecuteWithFallbackByTerminal(skillCtx.TerminalID, func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
			skillCtx.LLM = resolvedSet.LLM
			res, execErr := u.skill.Execute(skillCtx)
			if execErr == nil {
				result = res
			}
			return execErr
		})
	} else {
		// Use standard health-aware fallback
		err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
			skillCtx.LLM = resolvedSet.LLM
			res, execErr := u.skill.Execute(skillCtx)
			if execErr == nil {
				result = res
			}
			return execErr
		})
	}

	return result, err
}
