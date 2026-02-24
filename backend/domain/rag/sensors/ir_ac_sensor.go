package sensors

import (
	"fmt"
	"strconv"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type IRACsensor struct{}

func NewIRACsensor() DeviceSensor {
	return &IRACsensor{}
}

func (s *IRACsensor) CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool {
	return device.RemoteID != ""
}

func (s *IRACsensor) ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, executor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error) {
	promptLower := strings.ToLower(prompt)
	isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "matikan") || strings.Contains(promptLower, "mati")

	params, actionMsg := s.prepareACParams(isOff, promptLower, history)

	// Execute IR AC command
	success, err := executor.SendIRACCommand(token, device.ID, device.RemoteID, params)
	if err != nil {
		return &dtos.ControlResultDTO{
			Message:        fmt.Sprintf("Failed to execute command: %v", err),
			HTTPStatusCode: 500,
		}, nil
	}

	if !success {
		return &dtos.ControlResultDTO{
			Message:        "Command failed",
			HTTPStatusCode: 400,
		}, nil
	}

	return &dtos.ControlResultDTO{
		Message:  fmt.Sprintf("Successfully %s %s.", actionMsg, device.Name),
		DeviceID: device.ID,
	}, nil
}

// prepareACParams builds the parameters for an AC IR command, inheriting from history or using defaults (18°C) if needed.
func (s *IRACsensor) prepareACParams(isOff bool, promptLower string, history []string) (map[string]int, string) {
	params := make(map[string]int)
	var actions []string

	if isOff {
		params["power"] = 0
		return params, "turned off"
	}

	params["power"] = 1

	// 1. Search in current prompt
	temp, tempFound := s.parseTemperature(promptLower)
	mode, modeFound := s.parseMode(promptLower)
	wind, windFound := s.parseFanSpeed(promptLower)

	// 2. Search in history for missing values
	if !tempFound || !modeFound || !windFound {
		for i := len(history) - 1; i >= 0 && i >= len(history)-3; i-- {
			hLower := strings.ToLower(history[i])
			if !tempFound {
				if hTemp, found := s.parseTemperature(hLower); found {
					temp = hTemp
					tempFound = true
					utils.LogDebug("IRACsensor: Inherited temperature %d from history", temp)
				}
			}
			if !modeFound {
				if hMode, found := s.parseMode(hLower); found {
					mode = hMode
					modeFound = true
					utils.LogDebug("IRACsensor: Inherited mode %d from history", mode)
				}
			}
			if !windFound {
				if hWind, found := s.parseFanSpeed(hLower); found {
					wind = hWind
					windFound = true
					utils.LogDebug("IRACsensor: Inherited wind %d from history", wind)
				}
			}
		}
	}

	// 3. Apply Defaults if still missing (only apply if needed based on mode)
	if !modeFound {
		mode = 0 // Cool
		modeFound = true
		utils.LogDebug("IRACsensor: Using default mode %d (Cool)", mode)
	}

	// Apply temp default only if Cool (0) or Heat (1) mode
	if !tempFound && (mode == 0 || mode == 1) {
		temp = 18 // Default lowest typical safe temp
		tempFound = true
		utils.LogDebug("IRACsensor: Using default temperature %d", temp)
	}

	// Apply wind default only if not Humidity (4) or Auto (2) mode
	if !windFound && mode != 2 && mode != 4 {
		wind = 0 // Auto
		windFound = true
		utils.LogDebug("IRACsensor: Using default fan speed %d (Auto)", wind)
	}

	// 4. Build response actions based on mode requirements
	modeNames := map[int]string{0: "Cool", 1: "Heat", 2: "Auto", 3: "Wind", 4: "Humidity"}

	// Add mode to params
	if modeFound {
		params["mode"] = mode
		actions = append(actions, fmt.Sprintf("set mode to %s", modeNames[mode]))
	}

	// Modes: 0=Cool, 1=Heat, 2=Auto, 3=Fan/Wind, 4=Dry/Humidity
	// According to requirements:
	// - Humidity (4) and Auto (2): only mode + power, no temp/wind
	// - Fan (3): mode + power + wind, no temp
	// - Cool (0) and Heat (1): mode + power + temp + wind
	switch mode {
	case 2, 4: // Auto or Humidity
		// Only mode and power are needed
	case 3: // Fan/Wind
		if windFound {
			params["wind"] = wind
			windNames := map[int]string{0: "Auto", 1: "Low", 2: "Medium", 3: "High"}
			actions = append(actions, fmt.Sprintf("set fan speed to %s", windNames[wind]))
		}
	default: // Cool (0) or Heat (1)
		if tempFound {
			params["temp"] = temp
			actions = append(actions, fmt.Sprintf("set temperature to %d°C", temp))
		}
		if windFound {
			params["wind"] = wind
			windNames := map[int]string{0: "Auto", 1: "Low", 2: "Medium", 3: "High"}
			actions = append(actions, fmt.Sprintf("set fan speed to %s", windNames[wind]))
		}
	}

	if len(actions) == 0 {
		return params, "turned on"
	}

	return params, strings.Join(actions, ", ")
}

// parseTemperature extracts a temperature value (16-30) from a prompt.
func (s *IRACsensor) parseTemperature(promptLower string) (int, bool) {
	words := strings.Fields(promptLower)
	for i, word := range words {
		// Check for numeric literals
		if num, err := strconv.Atoi(word); err == nil && num >= 16 && num <= 30 {
			utils.LogDebug("IRACsensor: Found temperature value '%d' in prompt", num)
			return num, true
		}
		// Check for patterns like "ke 20" or "to 24"
		if (word == "ke" || word == "to") && i+1 < len(words) {
			if num, err := strconv.Atoi(words[i+1]); err == nil && num >= 16 && num <= 30 {
				utils.LogDebug("IRACsensor: Found temperature pattern '%s %d' in prompt", word, num)
				return num, true
			}
		}
	}
	return 0, false
}

// parseMode extracts an AC mode value from a prompt.
func (s *IRACsensor) parseMode(promptLower string) (int, bool) {
	// 0: Cool, 1: Heat, 2: Auto, 3: Fan/Wind, 4: Dry/Humidity

	// Priority 1: Explicit technical commands (must win over descriptive adjectives)
	if strings.Contains(promptLower, "auto") || strings.Contains(promptLower, "otomatis") {
		return 2, true
	}
	if strings.Contains(promptLower, "dry") || strings.Contains(promptLower, "humidity") || strings.Contains(promptLower, "lembab") || strings.Contains(promptLower, "kelembaban") {
		return 4, true
	}
	if strings.Contains(promptLower, "fan") || strings.Contains(promptLower, "wind") || strings.Contains(promptLower, "kipas") || strings.Contains(promptLower, "angin") {
		return 3, true
	}

	// Priority 2: Descriptive adjectives (can be used as mode indicators if no explicit mode is found)
	if strings.Contains(promptLower, "cool") || strings.Contains(promptLower, "dingin") {
		return 0, true
	}
	if strings.Contains(promptLower, "heat") || strings.Contains(promptLower, "panas") {
		return 1, true
	}

	return 0, false
}

// parseFanSpeed extracts an AC fan speed value from a prompt.
func (s *IRACsensor) parseFanSpeed(promptLower string) (int, bool) {
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
