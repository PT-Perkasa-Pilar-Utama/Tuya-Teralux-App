package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/skills"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type ControlUseCase interface {
	ProcessControl(uid, teraluxID, prompt string) (*dtos.ControlResultDTO, error)
}

type controlUseCase struct {
	llm          skills.LLMClient
	config       *utils.Config
	vector       *infrastructure.VectorService
	badger       *infrastructure.BadgerService
	tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor
	tuyaAuth     tuyaUsecases.TuyaAuthUseCase
}

func NewControlUseCase(llm skills.LLMClient, cfg *utils.Config, vector *infrastructure.VectorService, badger *infrastructure.BadgerService, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, tuyaAuth tuyaUsecases.TuyaAuthUseCase) ControlUseCase {
	return &controlUseCase{
		llm:          llm,
		config:       cfg,
		vector:       vector,
		badger:       badger,
		tuyaExecutor: tuyaExecutor,
		tuyaAuth:     tuyaAuth,
	}
}

func (u *controlUseCase) ProcessControl(uid, teraluxID, prompt string) (*dtos.ControlResultDTO, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	// Delegate control logic to ControlSkill
	skill := skills.NewControlSkill(u.tuyaExecutor, u.tuyaAuth)

	ctx := &skills.SkillContext{
		UID:       uid,
		TeraluxID: teraluxID,
		Prompt:    prompt,
		LLM:       u.llm,
		Config:    u.config,
		Vector:    u.vector,
		Badger:    u.badger,
	}

	// Preload history for the skill
	// This maintains the behavior where history is loaded from Badger
	historyKey := fmt.Sprintf("chat_history:%s", teraluxID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
		}
	}
	ctx.History = history

	res, err := skill.Execute(ctx)
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
