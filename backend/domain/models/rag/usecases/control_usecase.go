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
	tuyaUsecases "sensio/domain/tuya/usecases"
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
	tuyaExecutor     tuyaUsecases.TuyaDeviceControlExecutor
	tuyaAuth         tuyaUsecases.TuyaAuthUseCase
	skill            skills.Skill
	providerResolver providers.ProviderResolver
}

func NewControlUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, vector *infrastructure.VectorService, badger *infrastructure.BadgerService, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, tuyaAuth tuyaUsecases.TuyaAuthUseCase, skill skills.Skill, providerResolver providers.ProviderResolver) ControlUseCase {
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

	// Resolve provider based on terminal ID
	providerStart := time.Now()
	llmClient := u.llm
	fallbackClient := u.fallbackLLM
	if u.providerResolver != nil && terminalID != "" {
		resolveStart := time.Now()
		resolved, err := u.providerResolver.ResolveByTerminalID(terminalID)
		resolveDuration := time.Since(resolveStart)
		
		if err != nil {
			utils.LogWarn("ControlUseCase: Provider resolution failed for terminal %s | duration_ms=%d | error=%v, using default", terminalID, resolveDuration.Milliseconds(), err)
		} else {
			llmClient = resolved.LLM
			fallbackClient = resolved.FallbackLLM
			utils.LogInfo("ControlUseCase: Using terminal-specific provider '%s' for terminal %s | resolution_duration_ms=%d", resolved.ProviderName, terminalID, resolveDuration.Milliseconds())
		}
	}
	providerDuration := time.Since(providerStart)
	utils.LogDebug("ControlUseCase: Provider resolution completed | terminalID=%s | duration_ms=%d", terminalID, providerDuration.Milliseconds())

	skillCtx := &skills.SkillContext{
		Ctx:        ctx,
		UID:        uid,
		TerminalID: terminalID,
		Prompt:     prompt,
		LLM:        llmClient,
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
	res, err := u.skill.Execute(skillCtx)
	skillDuration := time.Since(skillStart)
	
	utils.LogDebug("ControlUseCase: Skill.Execute completed | duration_ms=%d | err=%v", skillDuration.Milliseconds(), err)
	
	if err != nil && fallbackClient != nil {
		utils.LogWarn("Control: Primary LLM failed | duration_ms=%d | error=%v, falling back to local model", skillDuration.Milliseconds(), err)
		fallbackStart := time.Now()
		skillCtx.LLM = fallbackClient
		res, err = u.skill.Execute(skillCtx)
		fallbackDuration := time.Since(fallbackStart)
		
		if err == nil {
			utils.LogInfo("Control: Fallback LLM succeeded | fallback_duration_ms=%d", fallbackDuration.Milliseconds())
		} else {
			utils.LogWarn("Control: Fallback LLM also failed | fallback_duration_ms=%d | error=%v", fallbackDuration.Milliseconds(), err)
		}
	}

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
