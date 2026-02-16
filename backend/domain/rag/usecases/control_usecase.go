package usecases

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
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

	// Determine command type from prompt
	promptLower := strings.ToLower(prompt)
	isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "turn off") || strings.Contains(promptLower, "matikan")
	
	// Check if this is an IR device (has remote_id)
	if target.RemoteID != "" {
		// IR AC device - use IR command API
		
		// Parse all parameters from current prompt
		temp, tempFound := u.parseTemperature(promptLower)
		mode, modeFound := u.parseMode(promptLower)
		wind, windFound := u.parseFanSpeed(promptLower)
		
		// Conditional Inheritance: Only check history for missing values
		if (!tempFound || !modeFound || !windFound) && u.badger != nil {
			historyKey := fmt.Sprintf("chat_history:%s", teraluxID)
			data, _ := u.badger.Get(historyKey)
			if data != nil {
				var history []string
				if err := json.Unmarshal(data, &history); err == nil && len(history) > 0 {
					// Check last few messages for missing commands
					for i := len(history) - 1; i >= 0 && i >= len(history)-3; i-- {
						hLower := strings.ToLower(history[i])
						
						if !tempFound {
							if hTemp, found := u.parseTemperature(hLower); found {
								temp = hTemp
								tempFound = true
								utils.LogDebug("ControlUseCase: Inherited temperature %d from history", temp)
							}
						}
						
						if !modeFound {
							if hMode, found := u.parseMode(hLower); found {
								mode = hMode
								modeFound = true
								utils.LogDebug("ControlUseCase: Inherited mode %d from history", mode)
							}
						}

						if !windFound {
							if hWind, found := u.parseFanSpeed(hLower); found {
								wind = hWind
								windFound = true
								utils.LogDebug("ControlUseCase: Inherited wind %d from history", wind)
							}
						}
					}
				}
			}
		}

		// Prepare command map
		params := make(map[string]int)
		var actions []string

		if isOff {
			params["power"] = 0
			actions = append(actions, "turned off")
		} else {
			params["power"] = 1
			if modeFound {
				params["mode"] = mode
				modeNames := map[int]string{0: "Cool", 1: "Heat", 2: "Auto", 3: "Wind", 4: "Humidity"}
				actions = append(actions, fmt.Sprintf("set mode to %s", modeNames[mode]))
			}
			if tempFound {
				params["temp"] = temp
				actions = append(actions, fmt.Sprintf("set temperature to %d°C", temp))
			}
			if windFound {
				params["wind"] = wind
				windNames := map[int]string{0: "Auto", 1: "Low", 2: "Medium", 3: "High"}
				actions = append(actions, fmt.Sprintf("set fan speed to %s", windNames[wind]))
			}
			if len(actions) == 0 {
				actions = append(actions, "turned on")
			}
		}

		action := strings.Join(actions, ", ")
		utils.LogDebug("ControlUseCase: Executing IR command - Params: %+v, Action: %s", params, action)
		
		success, err := u.tuyaExecutor.SendIRACCommand(token, target.ID, target.RemoteID, params)
		if err != nil {
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Failed to execute IR command for **%s**: %v", target.Name, err),
				DeviceID:       target.ID,
				HTTPStatusCode: 500,
			}, nil
		} else if success {
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Successfully %s **%s**.", action, target.Name),
				DeviceID:       target.ID,
				HTTPStatusCode: 200,
			}, nil
		}
		
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Command sent but execution status unclear for **%s**.", target.Name),
			DeviceID:       target.ID,
			HTTPStatusCode: 500,
		}, nil
	}
	
	// Regular switch device - detect switch code
	var commands []tuyaDtos.TuyaCommandDTO
	
	// Check if device has switch codes (switch_1, switch_2, etc.)
	for _, status := range target.Status {
		if strings.HasPrefix(status.Code, "switch_") || status.Code == "switch_led" {
			// Use the first available switch code
			commands = append(commands, tuyaDtos.TuyaCommandDTO{
				Code:  status.Code,
				Value: !isOff,
			})
			utils.LogDebug("ControlUseCase: Selected switch code '%s' for device '%s'", status.Code, target.Name)
			break
		}
	}
	
	// If still no command found, return error
	if len(commands) == 0 {
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Device **%s** does not support on/off control.", target.Name),
			DeviceID:       target.ID,
			HTTPStatusCode: 400,
		}, nil
	}

	success, err := u.tuyaExecutor.SendSwitchCommand(token, target.ID, commands)
	if err != nil {
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Failed to execute command for **%s**: %v", target.Name, err),
			DeviceID:       target.ID,
			HTTPStatusCode: 500,
		}, nil
	} else if success {
		action := "turned on"
		if isOff {
			action = "turned off"
		}
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Successfully %s **%s**.", action, target.Name),
			DeviceID:       target.ID,
			HTTPStatusCode: 200,
		}, nil
	}

	return &dtos.ControlResultDTO{
		Message:        fmt.Sprintf("Command sent but execution status unclear for **%s**.", target.Name),
		DeviceID:       target.ID,
		HTTPStatusCode: 500,
	}, nil
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

		// Determine command type from prompt
		promptLower := strings.ToLower(prompt)
		isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "turn off") || strings.Contains(promptLower, "matikan")
		
		// Check if this is an IR device (has remote_id)
		if targetDevice.RemoteID != "" {
			// IR AC device - use IR command API
			
			// 3. Handle AC parameters (mode, temp, wind)
			temp, tempFound := u.parseTemperature(promptLower)
			mode, modeFound := u.parseMode(promptLower)
			wind, windFound := u.parseFanSpeed(promptLower)
			
			// Inherit missing fields from history
			if !tempFound || !modeFound || !windFound {
				for i := len(history) - 1; i >= 0 && i >= len(history)-3; i-- {
					hLower := strings.ToLower(history[i])
					
					if !tempFound {
						if hTemp, found := u.parseTemperature(hLower); found {
							temp = hTemp
							tempFound = true
							utils.LogDebug("ControlUseCase: (Selection) Inherited temperature %d from history", temp)
						}
					}
					
					if !modeFound {
						if hMode, found := u.parseMode(hLower); found {
							mode = hMode
							modeFound = true
							utils.LogDebug("ControlUseCase: (Selection) Inherited mode %d from history", mode)
						}
					}

					if !windFound {
						if hWind, found := u.parseFanSpeed(hLower); found {
							wind = hWind
							windFound = true
							utils.LogDebug("ControlUseCase: (Selection) Inherited wind %d from history", wind)
						}
					}
				}
			}

			// Prepare command map
			params := make(map[string]int)
			var actions []string

			if isOff {
				params["power"] = 0
				actions = append(actions, "turned off")
			} else {
				params["power"] = 1
				if modeFound {
					params["mode"] = mode
					modeNames := map[int]string{0: "Cool", 1: "Heat", 2: "Auto", 3: "Wind", 4: "Humidity"}
					actions = append(actions, fmt.Sprintf("set mode to %s", modeNames[mode]))
				}
				if tempFound {
					params["temp"] = temp
					actions = append(actions, fmt.Sprintf("set temperature to %d°C", temp))
				}
				if windFound {
					params["wind"] = wind
					windNames := map[int]string{0: "Auto", 1: "Low", 2: "Medium", 3: "High"}
					actions = append(actions, fmt.Sprintf("set fan speed to %s", windNames[wind]))
				}
				if len(actions) == 0 {
					actions = append(actions, "turned on")
				}
			}

			action := strings.Join(actions, ", ")
			utils.LogDebug("ControlUseCase: (Selection) Executing IR command - Params: %+v, Action: %s", params, action)
			
			success, err := u.tuyaExecutor.SendIRACCommand(token, targetDevice.ID, targetDevice.RemoteID, params)
			if err != nil {
				return &dtos.ControlResultDTO{
					Message:        fmt.Sprintf("Failed to execute IR command for **%s**: %v", targetDevice.Name, err),
					DeviceID:       targetDevice.ID,
					HTTPStatusCode: 500,
				}, nil
			} else if success {
				return &dtos.ControlResultDTO{
					Message:        fmt.Sprintf("Successfully %s **%s**.", action, targetDevice.Name),
					DeviceID:       targetDevice.ID,
					HTTPStatusCode: 200,
				}, nil
			}
			
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Command sent but execution status unclear for **%s**.", targetDevice.Name),
				DeviceID:       targetDevice.ID,
				HTTPStatusCode: 500,
			}, nil
		}
		
		// Regular switch device - detect switch code
		var commands []tuyaDtos.TuyaCommandDTO
		
		// Check if device has switch codes
		for _, status := range targetDevice.Status {
			if strings.HasPrefix(status.Code, "switch_") || status.Code == "switch_led" {
				commands = append(commands, tuyaDtos.TuyaCommandDTO{
					Code:  status.Code,
					Value: !isOff,
				})
				break
			}
		}
		
		// If still no command found, return error
		if len(commands) == 0 {
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Device **%s** does not support on/off control.", targetDevice.Name),
				DeviceID:       targetDevice.ID,
				HTTPStatusCode: 400,
			}, nil
		}

		success, err := u.tuyaExecutor.SendSwitchCommand(token, targetDevice.ID, commands)
		if err != nil {
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Failed to execute command for **%s**: %v", targetDevice.Name, err),
				DeviceID:       targetDevice.ID,
				HTTPStatusCode: 500,
			}, nil
		} else if success {
			action := "turned on"
			if isOff {
				action = "turned off"
			}
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Successfully %s **%s**.", action, targetDevice.Name),
				DeviceID:       targetDevice.ID,
				HTTPStatusCode: 200,
			}, nil
		}

		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Command sent but execution status unclear for **%s**.", targetDevice.Name),
			DeviceID:       targetDevice.ID,
			HTTPStatusCode: 500,
		}, nil
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

// parseTemperature extracts a temperature value (16-30) from a prompt.
func (u *controlUseCase) parseTemperature(promptLower string) (int, bool) {
	words := strings.Fields(promptLower)
	for i, word := range words {
		// Check for numeric literals
		if num, err := strconv.Atoi(word); err == nil && num >= 16 && num <= 30 {
			utils.LogDebug("ControlUseCase: Found temperature value '%d' in prompt", num)
			return num, true
		}
		// Check for patterns like "ke 20" or "to 24"
		if (word == "ke" || word == "to") && i+1 < len(words) {
			if num, err := strconv.Atoi(words[i+1]); err == nil && num >= 16 && num <= 30 {
				utils.LogDebug("ControlUseCase: Found temperature pattern '%s %d' in prompt", word, num)
				return num, true
			}
		}
	}
	return 0, false
}

// parseMode extracts an AC mode value from a prompt.
func (u *controlUseCase) parseMode(promptLower string) (int, bool) {
	// 0: Cool, 1: Heat, 2: Auto, 3: Fan/Wind, 4: Dry/Humidity
	
	if strings.Contains(promptLower, "cool") || strings.Contains(promptLower, "dingin") {
		return 0, true
	}
	if strings.Contains(promptLower, "heat") || strings.Contains(promptLower, "panas") {
		return 1, true
	}
	if strings.Contains(promptLower, "auto") || strings.Contains(promptLower, "otomatis") {
		return 2, true
	}
	if strings.Contains(promptLower, "fan") || strings.Contains(promptLower, "wind") || strings.Contains(promptLower, "kipas") || strings.Contains(promptLower, "angin") {
		return 3, true
	}
	if strings.Contains(promptLower, "dry") || strings.Contains(promptLower, "humidity") || strings.Contains(promptLower, "lembab") || strings.Contains(promptLower, "kelembaban") {
		return 4, true
	}
	
	return 0, false
}

// parseFanSpeed extracts an AC fan speed value from a prompt.
func (u *controlUseCase) parseFanSpeed(promptLower string) (int, bool) {
	// 0: Auto, 1: Low, 2: Med, 3: High
	
	if strings.Contains(promptLower, "low") || strings.Contains(promptLower, "pelan") || strings.Contains(promptLower, "kecil") {
		return 1, true
	}
	if strings.Contains(promptLower, "medium") || strings.Contains(promptLower, "sedang") {
		return 2, true
	}
	if strings.Contains(promptLower, "high") || strings.Contains(promptLower, "kencang") || strings.Contains(promptLower, "cepat") || strings.Contains(promptLower, "besar") {
		return 3, true
	}
	if strings.Contains(promptLower, "auto") || strings.Contains(promptLower, "otomatis") {
		return 0, true
	}
	
	return 0, false
}
