package sensors

import (
	"fmt"
	"strconv"
	"strings"
	"teralux_app/domain/rag/dtos"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type DeviceSensor interface {
	CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool
	ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error)
}

type TeraluxSensor struct{}

func NewTeraluxSensor() DeviceSensor {
	return &TeraluxSensor{}
}

func (s *TeraluxSensor) CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool {
	// Handle Teralux devices with voice/media controls (dgnzk category)
	return device.Category == "dgnzk"
}

func (s *TeraluxSensor) ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, executor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error) {
	promptLower := strings.ToLower(prompt)

	// Check what to control
	var commands []tuyaDtos.TuyaCommandDTO
	var actionMsg string

	// Try to match device-specific codes (voice_mic, play, volume_set, brightness, work_mode, etc)
	if code, value, matched := s.matchDeviceCode(promptLower, device.Status); matched {
		commands = append(commands, tuyaDtos.TuyaCommandDTO{
			Code:  code,
			Value: value,
		})
		actionMsg = fmt.Sprintf("controlled %s", code)
	} else {
		// Default switch control - support all switch variants
		isOn := strings.Contains(promptLower, "on") || strings.Contains(promptLower, "nyalakan") || strings.Contains(promptLower, "hidupkan")
		isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "matikan") || strings.Contains(promptLower, "mati")

		if !isOn && !isOff {
			isOn = true // Default to ON
		}

		// Find any switch code - support: switch, switch1, switch_1, switch_led, switch_2, switch_3
		switchCode, switchIndex := s.findSwitchCode(promptLower, device.Status)

		if switchCode == "" {
			return &dtos.ControlResultDTO{
				Message:        fmt.Sprintf("Device %s does not have a switchable control.", device.Name),
				HTTPStatusCode: 400,
			}, nil
		}

		commands = append(commands, tuyaDtos.TuyaCommandDTO{
			Code:  switchCode,
			Value: isOn,
		})

		if isOn {
			actionMsg = "turned on"
		} else {
			actionMsg = "turned off"
		}

		// Add switch index to message if multiple switches
		if switchIndex > 0 {
			actionMsg = fmt.Sprintf("%s (switch %d)", actionMsg, switchIndex)
		}
	}

	// Execute command
	success, err := executor.SendSwitchCommand(token, device.ID, commands)
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

// findSwitchCode finds any switch code in the device status
// Returns the switch code and switch index (1-based, 0 if single switch)
func (s *TeraluxSensor) findSwitchCode(promptLower string, status []tuyaDtos.TuyaDeviceStatusDTO) (string, int) {
	// Check for specific switch number in prompt
	switchNum := s.extractSwitchNumber(promptLower)

	// Collect all switch codes
	var switchCodes []string
	for _, stat := range status {
		if strings.HasPrefix(stat.Code, "switch") {
			switchCodes = append(switchCodes, stat.Code)
		}
	}

	if len(switchCodes) == 0 {
		return "", 0
	}

	// If specific switch requested
	if switchNum > 0 {
		// Try to find switch_N or switchN format
		targetCodes := []string{
			fmt.Sprintf("switch_%d", switchNum),
			fmt.Sprintf("switch%d", switchNum),
		}
		for _, targetCode := range targetCodes {
			for _, code := range switchCodes {
				if code == targetCode {
					return code, switchNum
				}
			}
		}
	}

	// Return first switch found
	return switchCodes[0], 0
}

// extractSwitchNumber extracts switch number from prompt (e.g., "switch 1" -> 1)
func (s *TeraluxSensor) extractSwitchNumber(promptLower string) int {
	// Look for patterns like "switch 1", "switch1", "saklar 1", etc.
	patterns := []string{"switch ", "saklar ", "lampu "}
	for _, pattern := range patterns {
		if idx := strings.Index(promptLower, pattern); idx != -1 {
			remaining := promptLower[idx+len(pattern):]
			words := strings.Fields(remaining)
			if len(words) > 0 {
				if num, err := strconv.Atoi(words[0]); err == nil && num >= 1 && num <= 10 {
					return num
				}
			}
		}
	}
	return 0
}

// matchDeviceCode tries to match a user prompt to an available device control code
func (s *TeraluxSensor) matchDeviceCode(promptLower string, status []tuyaDtos.TuyaDeviceStatusDTO) (string, interface{}, bool) {
	// Map keywords to device codes for Teralux voice/media controls

	isOn := strings.Contains(promptLower, "on") || strings.Contains(promptLower, "nyalakan") || strings.Contains(promptLower, "hidupkan")
	isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "matikan") || strings.Contains(promptLower, "mati")

	actionValue := true
	if isOff && !isOn {
		actionValue = false
	}

	// Voice/Microphone controls
	if strings.Contains(promptLower, "voice mic") || strings.Contains(promptLower, "voice_mic") ||
		strings.Contains(promptLower, "mikrofon") || strings.Contains(promptLower, "mic") {
		for _, s := range status {
			if s.Code == "voice_mic" {
				// If specific action provided, follow it, otherwise toggle
				var val bool
				if isOn || isOff {
					val = actionValue
				} else {
					val = !parseBoolean(s.Value)
				}
				return s.Code, val, true
			}
		}
	}

	// Play/Playback controls
	if strings.Contains(promptLower, "play") || strings.Contains(promptLower, "putar") ||
		strings.Contains(promptLower, "musik") || strings.Contains(promptLower, "music") {
		for _, s := range status {
			if s.Code == "play" || s.Code == "voice_play" {
				return s.Code, actionValue, true
			}
		}
	}

	// Volume controls
	if strings.Contains(promptLower, "volume") || strings.Contains(promptLower, "volume_set") ||
		strings.Contains(promptLower, "keras") {
		// Check for numeric volume value
		if vol, found := s.parseVolumeValue(promptLower); found {
			for _, s := range status {
				if s.Code == "volume_set" || s.Code == "voice_vol" {
					return s.Code, vol, true
				}
			}
		}
		// Default to medium volume if "volume" mentioned but no specific value
		for _, s := range status {
			if s.Code == "volume_set" || s.Code == "voice_vol" {
				return s.Code, 50, true
			}
		}
	}

	return "", nil, false
}

// parseBoolean converts device value to bool
func parseBoolean(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return strings.ToLower(v) == "true" || v == "1"
	}
	return false
}

// parseVolumeValue extracts volume percentage (0-100) from prompt
func (s *TeraluxSensor) parseVolumeValue(promptLower string) (int, bool) {
	words := strings.Fields(promptLower)
	for i, word := range words {
		if num, err := strconv.Atoi(word); err == nil && num >= 0 && num <= 100 {
			// Check if this looks like a volume value (not temperature or other number)
			if i > 0 {
				prevWord := words[i-1]
				if strings.Contains(prevWord, "volume") || strings.Contains(prevWord, "keras") ||
					strings.Contains(promptLower, "volume") {
					return num, true
				}
			}
		}
	}
	return 0, false
}
