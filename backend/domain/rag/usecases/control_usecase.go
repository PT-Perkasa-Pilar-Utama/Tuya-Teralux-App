package usecases

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/skills"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"strings"
)

type ControlUseCase interface {
	ProcessControl(uid, terminalID, prompt string) (*dtos.ControlResultDTO, error)
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

func (u *controlUseCase) ProcessControl(uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	// Delegate control logic to ControlSkill
	if u.skill == nil {
		return nil, fmt.Errorf("control skill not configured")
	}

	ctx := &skills.SkillContext{
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
	ctx.History = history

	res, err := u.skill.Execute(ctx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("Control: Primary LLM failed, falling back to local model: %v", err)
		ctx.LLM = u.fallbackLLM
		res, err = u.skill.Execute(ctx)
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
