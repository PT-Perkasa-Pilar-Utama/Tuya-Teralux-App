package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/skills"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"strings"
)

type ControlUseCase interface {
	ProcessControl(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error)
}

type controlUseCase struct {
	llm          skills.LLMClient
	fallbackLLM  skills.LLMClient
	config       *utils.Config
	vector       *infrastructure.VectorService
	badger       *infrastructure.BadgerService
	tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor
	tuyaAuth     tuyaUsecases.TuyaAuthUseCase
	skill        skills.Skill
}

func NewControlUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *utils.Config, vector *infrastructure.VectorService, badger *infrastructure.BadgerService, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, tuyaAuth tuyaUsecases.TuyaAuthUseCase, skill skills.Skill) ControlUseCase {
	return &controlUseCase{
		llm:          llm,
		fallbackLLM:  fallbackLLM,
		config:       cfg,
		vector:       vector,
		badger:       badger,
		tuyaExecutor: tuyaExecutor,
		tuyaAuth:     tuyaAuth,
		skill:        skill,
	}
}

func (u *controlUseCase) ProcessControl(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
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

	skillCtx := &skills.SkillContext{
		Ctx:        ctx,
		UID:        uid,
		TerminalID: terminalID,
		Prompt:     prompt,
		LLM:        u.llm,
		Config:     u.config,
		Vector:     u.vector,
		Badger:     u.badger,
	}

	// Preload history for the skill
	// This maintains the behavior where history is loaded from Badger
	historyKey := fmt.Sprintf("chat_history:%s", terminalID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
		}
	}
	skillCtx.History = history

	res, err := u.skill.Execute(skillCtx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("Control: Primary LLM failed, falling back to local model: %v", err)
		skillCtx.LLM = u.fallbackLLM
		res, err = u.skill.Execute(skillCtx)
	}

	if err != nil {
		return nil, err
	}

	deviceID := ""
	if dataMap, ok := res.Data.(map[string]interface{}); ok {
		if id, ok := dataMap["device_id"].(string); ok {
			deviceID = id
		}
	}

	return &dtos.ControlResultDTO{
		Message:        res.Message,
		DeviceID:       deviceID,
		HTTPStatusCode: res.HTTPStatusCode,
	}, nil

}
