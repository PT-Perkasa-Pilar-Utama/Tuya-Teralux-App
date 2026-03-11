package sensors

import (
	"fmt"
	"regexp"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	tuyaDtos "sensio/domain/tuya/dtos"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"strconv"
	"strings"
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

	params, actionMsg, customRes := s.prepareACParams(isOff, promptLower, history)
	if customRes != nil {
		return customRes, nil
	}

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
func (s *IRACsensor) prepareACParams(isOff bool, promptLower string, history []string) (map[string]int, string, *dtos.ControlResultDTO) {
	params := make(map[string]int)
	var actions []string

	if isOff {
		params["power"] = 0
		return params, "turned off", nil
	}

	params["power"] = 1

	// 1. Search in current prompt
	temp, tempFound := s.parseTemperature(promptLower)
	mode, modeFound := s.parseMode(promptLower)
	wind, windFound := s.parseFanSpeed(promptLower)

	// Validation: Temperature Range (16-30)
	if tempFound {
		if temp < 16 || temp > 30 {
			return nil, "", &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Temperature %d is out of range. Valid range is 16-30°C.", temp),
				HTTPStatusCode: 400,
			}
		}
	}

	// 2. Search in history for missing values
	if !tempFound || !modeFound || !windFound {
		for i := len(history) - 1; i >= 0 && i >= len(history)-6; i-- {
			// Skip assistant messages to avoid context poisoning from its own descriptions (e.g., "kecepatan", "otomatis")
			if strings.HasPrefix(history[i], "Assistant:") {
				continue
			}

			hLower := strings.ToLower(history[i])
			if !tempFound {
				if hTemp, found := s.parseTemperature(hLower); found {
					// Historical temperature also needs to be valid
					if hTemp >= 16 && hTemp <= 30 {
						temp = hTemp
						tempFound = true
						utils.LogDebug("IRACsensor: Inherited temperature %d from history", temp)
					}
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

	// Override inherited or prompt mode if user explicitly provides temperature in the prompt
	// but the mode (e.g. from history or explicitly in prompt) does not support manual temperature.
	_, tempInPrompt := s.parseTemperature(promptLower)

	if tempInPrompt && modeFound {
		if mode == 2 || mode == 3 || mode == 4 { // Auto, Wind, Humidity
			mode = 0 // Override to Cool
			utils.LogDebug("IRACsensor: Explicit temperature requested, overriding mode to %d (Cool)", mode)
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
		return params, "turned on", nil
	}

	return params, strings.Join(actions, ", "), nil
}

// parseTemperature extracts a temperature value from a prompt using context-aware logic.
func (s *IRACsensor) parseTemperature(prompt string) (int, bool) {
	promptLower := strings.ToLower(prompt)
	// 1. Clean and tokenize
	// Insert spaces around numbers to handle "25c", "25°C", etc.
	reSpace := regexp.MustCompile(`(\d+)`)
	spaced := reSpace.ReplaceAllString(promptLower, " $1 ")

	cleaned := strings.ReplaceAll(spaced, "°c", " degc ")
	cleaned = strings.ReplaceAll(cleaned, "derajat", " deg ")
	cleaned = strings.ReplaceAll(cleaned, "degree", " deg ")
	cleaned = strings.ReplaceAll(cleaned, ".", " ")
	cleaned = strings.ReplaceAll(cleaned, ",", " ")
	tokens := strings.Fields(cleaned)

	type candidate struct {
		val      int
		priority int // 3: high (keyword), 2: medium (intent), 1: fallback, 0: forbidden
	}

	var candidates []candidate

	tempKeywords := map[string]bool{
		"suhu": true, "temperatur": true, "temperature": true, "temp": true,
		"deg": true, "degc": true, "c": true, "celsius": true,
	}
	intentKeywords := map[string]bool{
		"ke": true, "to": true, "set": true, "jadi": true,
		"naikkan": true, "turunkan": true, "di": true,
	}
	forbiddenKeywords := map[string]bool{
		"menit": true, "jam": true, "detik": true, "second": true, "minute": true, "hour": true,
		"device": true, "no": true, "nomor": true, "id": true,
	}

	for i, token := range tokens {
		val, err := strconv.Atoi(token)
		if err != nil {
			continue
		}

		// Only consider numbers in typical AC range (10-40) to reduce noise
		if val < 10 || val > 40 {
			continue
		}

		prio := 1 // Default fallback priority

		// Check context (window of 2 words before and after)
		isForbidden := false
		isHigh := false
		isMedium := false

		start := i - 2
		if start < 0 {
			start = 0
		}
		end := i + 2
		if end >= len(tokens) {
			end = len(tokens) - 1
		}

		for j := start; j <= end; j++ {
			if i == j {
				continue
			}
			t := tokens[j]
			if forbiddenKeywords[t] {
				isForbidden = true
				break
			}
			if tempKeywords[t] {
				isHigh = true
			}
			if intentKeywords[t] {
				isMedium = true
			}
		}

		if isForbidden {
			prio = 0
		} else if isHigh {
			prio = 3
		} else if isMedium {
			prio = 2
		}

		candidates = append(candidates, candidate{val: val, priority: prio})
	}

	if len(candidates) == 0 {
		return 0, false
	}

	// Selection Strategy:
	// 1. Pick the one with the highest priority.
	// 2. If priorities are equal, pick the LAST one (common in "from X to Y" prompts).
	maxPrio := -1
	for _, c := range candidates {
		if c.priority > maxPrio {
			maxPrio = c.priority
		}
	}

	// Forbidden candidates (prio 0) are only allowed if they are the ONLY candidates
	// and we are feeling lucky? No, forbidden should be ignored unless priority 1+ exists.
	if maxPrio == 0 {
		return 0, false
	}

	var result int
	for _, c := range candidates {
		if c.priority == maxPrio {
			result = c.val
		}
	}

	utils.LogDebug("IRACsensor: Extracted temperature %d (prio %d) from prompt", result, maxPrio)
	return result, true
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
	// Note: avoided "cepat" as it matches "kecepatan", which causes false positives for High speed
	if strings.Contains(promptLower, "high") || strings.Contains(promptLower, "kencang") || strings.Contains(promptLower, "maksimal") || strings.Contains(promptLower, "besar") {
		return 3, true
	}
	if strings.Contains(promptLower, "auto") || strings.Contains(promptLower, "otomatis") {
		return 0, true
	}

	return 0, false
}
