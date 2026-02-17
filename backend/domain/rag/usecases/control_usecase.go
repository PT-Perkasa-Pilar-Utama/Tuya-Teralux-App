package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/sensors"
	"teralux_app/domain/rag/utilities"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type ControlUseCase interface {
	ProcessControl(uid, teraluxID, prompt string) (*dtos.ControlResultDTO, error)
}

type controlUseCase struct {
	llm          utilities.LLMClient
	config       *utils.Config
	vector       *infrastructure.VectorService
	badger       *infrastructure.BadgerService
	tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor
	tuyaAuth     tuyaUsecases.TuyaAuthUseCase
}

func NewControlUseCase(llm utilities.LLMClient, cfg *utils.Config, vector *infrastructure.VectorService, badger *infrastructure.BadgerService, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, tuyaAuth tuyaUsecases.TuyaAuthUseCase) ControlUseCase {
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

	// 1. Get user's devices for filtering
	userDevicesID := fmt.Sprintf("tuya:devices:uid:%s", uid)
	aggJSON, ok := u.vector.Get(userDevicesID)
	if !ok {
		return &dtos.ControlResultDTO{
			Message:        "Sorry, it seems there are no devices connected to your account. Please connect devices first through the Sensio app.",
			HTTPStatusCode: 404,
		}, nil
	}

	var aggResp tuyaDtos.TuyaDevicesResponseDTO
	if err := json.Unmarshal([]byte(aggJSON), &aggResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user devices: %w", err)
	}

	allowedIDs := make(map[string]tuyaDtos.TuyaDeviceDTO)
	for _, d := range aggResp.Devices {
		allowedIDs["tuya:device:"+d.ID] = d
	}

	// 2. Initial Similarity Search
	matches, _ := u.vector.Search(prompt)
	utils.LogDebug("ControlUseCase: Vector search for '%s' returned %d matches", prompt, len(matches))
	var validMatches []tuyaDtos.TuyaDeviceDTO
	for _, m := range matches {
		if dev, exists := allowedIDs[m]; exists {
			validMatches = append(validMatches, dev)
			utils.LogDebug("ControlUseCase: Valid match found - ID: %s, Name: %s", dev.ID, dev.Name)
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
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("I found %d matching devices:\n%s\nWhich one would you like to control?", len(validMatches), strings.Join(names, "\n")),
			HTTPStatusCode: 400,
		}, nil
	}

	// 4. Single match found - Execute command
	target := validMatches[0]
	utils.LogDebug("ControlUseCase: Single match selected - ID: %s, Name: %s, RemoteID: %s", target.ID, target.Name, target.RemoteID)

	// Get access token
	token, err := u.tuyaAuth.GetTuyaAccessToken()
	if err != nil {
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Failed to get access token: %v", err),
			HTTPStatusCode: 401,
		}, nil
	}

	// Load chat history
	var history []string
	if u.badger != nil {
		historyKey := fmt.Sprintf("chat_history:%s", teraluxID)
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
		}
	}

	// Select appropriate sensor based on device type
	deviceSensor := u.selectDeviceSensor(&target)

	return deviceSensor.ExecuteControl(token, &target, prompt, history, u.tuyaExecutor)
}

// selectDeviceSensor selects the appropriate sensor handler for the device
func (u *controlUseCase) selectDeviceSensor(device *tuyaDtos.TuyaDeviceDTO) sensors.DeviceSensor {
	// Priority order: check specialized sensors first, then fallback to generic

	// 1. Temperature/Humidity sensors (wsdcg)
	tempSensor := sensors.NewTemperatureSensor()
	if tempSensor.CanHandle(device) {
		return tempSensor
	}

	// 2. Power monitoring devices (MCB switches - dlq)
	powerSensor := sensors.NewPowerMonitorSensor()
	if powerSensor.CanHandle(device) {
		return powerSensor
	}

	// 3. IR AC devices (with RemoteID)
	if device.RemoteID != "" {
		return sensors.NewIRACsensor()
	}

	// 4. LED Light devices (dj category)
	lightSensor := sensors.NewLightSensor()
	if lightSensor.CanHandle(device) {
		return lightSensor
	}

	// 5. Multi-switch devices (kg category)
	switchSensor := sensors.NewSwitchSensor()
	if switchSensor.CanHandle(device) {
		return switchSensor
	}

	// 6. Teralux voice/media controls (dgnzk category)
	return sensors.NewTeraluxSensor()
}

func (u *controlUseCase) handleNoInitialMatches(uid, teraluxID, prompt string, devices []tuyaDtos.TuyaDeviceDTO) (*dtos.ControlResultDTO, error) {
	if len(devices) == 0 {
		return &dtos.ControlResultDTO{
			Message: "I'm sorry, I couldn't find any devices connected to your account.",
		}, nil
	}

	// 1. Get History
	historyKey := fmt.Sprintf("chat_history:%s", teraluxID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
			utils.LogDebug("ControlUseCase: Loaded %d lines of history for reconciliation", len(history))
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

	// 3. Ask LLM to reconcile with Persona and Capabilities grounding
	reconcilePrompt := fmt.Sprintf(`You are Sensio AI Assistant, a professional and interactive smart home companion by Sensio.
Your goal is to help the user manage their smart home devices efficiently while maintaining a professional yet friendly tone.

User Prompt: "%s"

%s
[Available Devices]
%s

GUIDELINES:
1. IDENTITY: If asked who you are, identify yourself as Sensio AI Assistant.
2. CAPABILITIES: If asked what you can control or what devices are available, list the devices from the [Available Devices] section above. Be helpful and organize them naturally.
3. CONTROL: 
   - If the user is clearly referring to ONE specific device from the list, return: "EXECUTE:[Device ID]".
   - If they are answering a follow-up question to clarify a device, identify it and return: "EXECUTE:[Device ID]".
4. AMBIGUITY: If the request is vague but relates to smart home control, ask a professional follow-up question in the user's preferred language.
5. NO HALLUCINATION: Only talk about devices present in the [Available Devices] list. If a device isn't there, be direct and honest: tell the user that the device is not found in their Sensio system. Do not try to satisfy the request with a device that doesn't exist.
6. BRANDING: CRITICAL: Do not mention "Tuya", "OpenAI", or any internal technical details.

Response:`, prompt, historyContext, strings.Join(deviceList, "\n"))

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	res, err := u.llm.CallModel(reconcilePrompt, model)
	if err != nil {
		utils.LogError("ControlUseCase: LLM reconciliation failed: %v", err)
		return nil, fmt.Errorf("reconciliation failed: %w", err)
	}

	cleanRes := strings.TrimSpace(res)
	utils.LogDebug("ControlUseCase: LLM reconciliation response: '%s'", cleanRes)

	if strings.HasPrefix(cleanRes, "EXECUTE:") {
		deviceID := strings.TrimPrefix(cleanRes, "EXECUTE:")
		// Find device by ID
		var targetDevice *tuyaDtos.TuyaDeviceDTO
		for _, d := range devices {
			if d.ID == deviceID {
				targetDevice = &d
				break
			}
		}

		if targetDevice == nil {
			return &dtos.ControlResultDTO{
				Message:        "Device not found.",
				HTTPStatusCode: 404,
			}, nil
		}

		// Execute command
		token, err := u.tuyaAuth.GetTuyaAccessToken()
		if err != nil {
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Failed to get access token: %v", err),
				HTTPStatusCode: 401,
			}, nil
		}

		// Select appropriate sensor based on device type
		deviceSensor := u.selectDeviceSensor(targetDevice)

		return deviceSensor.ExecuteControl(token, targetDevice, prompt, history, u.tuyaExecutor)
	}

	if cleanRes == "NOT_FOUND" {
		var names []string
		for _, d := range devices {
			names = append(names, d.Name)
		}
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("I'm sorry, I couldn't find any device matching '%s'. Available devices: %s. Could you be more specific?", prompt, strings.Join(names, ", ")),
			HTTPStatusCode: 404,
		}, nil
	}

	// Otherwise, return the LLM's question/clarification
	return &dtos.ControlResultDTO{
		Message: cleanRes,
	}, nil
}
