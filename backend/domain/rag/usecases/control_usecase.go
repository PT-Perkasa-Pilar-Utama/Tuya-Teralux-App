package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/utilities"
	tuyaDtos "teralux_app/domain/tuya/dtos"
)

type ControlUseCase interface {
	ProcessControl(uid, teraluxID, prompt string) (string, error)
}

type controlUseCase struct {
	llm    utilities.LLMClient
	config *utils.Config
	vector *infrastructure.VectorService
	badger *infrastructure.BadgerService
}

func NewControlUseCase(llm utilities.LLMClient, cfg *utils.Config, vector *infrastructure.VectorService, badger *infrastructure.BadgerService) ControlUseCase {
	return &controlUseCase{
		llm:    llm,
		config: cfg,
		vector: vector,
		badger: badger,
	}
}

func (u *controlUseCase) ProcessControl(uid, teraluxID, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	// 1. Get user's devices for filtering
	userDevicesID := fmt.Sprintf("tuya:devices:uid:%s", uid)
	aggJSON, ok := u.vector.Get(userDevicesID)
	if !ok {
		return "Maaf, sepertinya tidak ada perangkat yang terhubung ke akun Anda. Silakan hubungkan perangkat terlebih dahulu melalui aplikasi Sensio.", nil
	}

	var aggResp tuyaDtos.TuyaDevicesResponseDTO
	if err := json.Unmarshal([]byte(aggJSON), &aggResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal user devices: %w", err)
	}

	allowedIDs := make(map[string]tuyaDtos.TuyaDeviceDTO)
	for _, d := range aggResp.Devices {
		allowedIDs["tuya:device:"+d.ID] = d
	}

	// 2. Initial Similarity Search
	matches, _ := u.vector.Search(prompt)
	var validMatches []tuyaDtos.TuyaDeviceDTO
	for _, m := range matches {
		if dev, exists := allowedIDs[m]; exists {
			validMatches = append(validMatches, dev)
		}
	}

	// 3. Handle Ambiguity logic
	if len(validMatches) == 0 {
		// Try to broaden search or check if it's a follow-up answer using LLM
		return u.handleNoInitialMatches(uid, teraluxID, prompt, aggResp.Devices)
	}

	if len(validMatches) > 1 {
		// Multiple matches found, ask for clarification with specific names
		var names []string
		for _, v := range validMatches {
			names = append(names, fmt.Sprintf("- **%s**", v.Name))
		}
		return fmt.Sprintf("I found %d matching devices:\n%s\nWhich one would you like to control?", len(validMatches), strings.Join(names, "\n")), nil
	}

	// 4. Single match found
	target := validMatches[0]
	return fmt.Sprintf("Sure! Running command for **%s**.", target.Name), nil
}

func (u *controlUseCase) handleNoInitialMatches(uid, teraluxID, prompt string, devices []tuyaDtos.TuyaDeviceDTO) (string, error) {
	if len(devices) == 0 {
		return "I'm sorry, I couldn't find any devices connected to your account.", nil
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

	historyContext := ""
	if len(history) > 0 {
		historyContext = "Previous conversation:\n" + strings.Join(history, "\n") + "\n"
	}

	// 2. Build a list of searchable device names
	var deviceList []string
	for _, d := range devices {
		deviceList = append(deviceList, fmt.Sprintf("- %s (ID: %s)", d.Name, d.ID))
	}

	// 3. Ask LLM to reconcile (System Prompt in English only)
	reconcilePrompt := fmt.Sprintf(`You are a Smart Home Assistant. The user's input did not directly match any device name in a similarity search.
Determine if the user's input is a specific selection from the available devices, possibly based on previous conversation context.

User Prompt: "%s"

%s
Available Devices:
%s

If the user is clearly referring to ONE specific device from the list, return: "EXECUTE:[Device ID]".
If they are answering a follow-up question, identify the correct device.
If it's still ambiguous, return a helpful question in the user's preferred language asking them to choose from the list.
If it's neither, return: "NOT_FOUND".

Response:`, prompt, historyContext, strings.Join(deviceList, "\n"))

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	res, err := u.llm.CallModel(reconcilePrompt, model)
	if err != nil {
		return "", fmt.Errorf("reconciliation failed: %w", err)
	}

	cleanRes := strings.TrimSpace(res)
	if strings.HasPrefix(cleanRes, "EXECUTE:") {
		deviceID := strings.TrimPrefix(cleanRes, "EXECUTE:")
		// Find device name by ID
		for _, d := range devices {
			if d.ID == deviceID {
				return fmt.Sprintf("Sure! Running command for **%s**.", d.Name), nil
			}
		}
	}

	if cleanRes == "NOT_FOUND" {
		var names []string
		for _, d := range devices {
			names = append(names, d.Name)
		}
		return fmt.Sprintf("I'm sorry, I couldn't find any device matching '%s'. Available devices: %s. Could you be more specific?", prompt, strings.Join(names, ", ")), nil
	}

	// Otherwise, return the LLM's question/clarification
	return cleanRes, nil
}
