package sensors

import (
	"fmt"
	"strconv"
	"strings"
	"teralux_app/domain/rag/dtos"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type SwitchSensor struct{}

func NewSwitchSensor() DeviceSensor {
	return &SwitchSensor{}
}

func (s *SwitchSensor) CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool {
	// Handle multi-switch devices (kg category)
	if device.Category == "kg" {
		return true
	}

	// Also handle devices with multiple switch status codes
	switchCount := 0
	for _, status := range device.Status {
		if strings.HasPrefix(status.Code, "switch_") {
			switchCount++
		}
	}
	return switchCount >= 2
}

func (s *SwitchSensor) ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, executor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error) {
	promptLower := strings.ToLower(prompt)

	var commands []tuyaDtos.TuyaCommandDTO
	var actionMsg string

	isOn := strings.Contains(promptLower, "on") || strings.Contains(promptLower, "nyalakan") || strings.Contains(promptLower, "hidupkan")
	isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "matikan") || strings.Contains(promptLower, "mati")

	if !isOn && !isOff {
		isOn = true // Default to ON
	}

	// Extract which switch number from prompt
	switchNum := s.extractSwitchNumber(promptLower)

	// Find the target switch code
	var switchCode string
	if switchNum > 0 {
		// Look for specific switch
		targetCode := fmt.Sprintf("switch_%d", switchNum)
		for _, status := range device.Status {
			if status.Code == targetCode {
				switchCode = targetCode
				break
			}
		}
	}

	// If no specific switch found, use first available switch
	if switchCode == "" {
		for _, status := range device.Status {
			if strings.HasPrefix(status.Code, "switch_") {
				switchCode = status.Code
				break
			}
		}
	}

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

	// Add switch number to message if multiple switches
	if switchNum > 0 {
		actionMsg = fmt.Sprintf("%s switch %d", actionMsg, switchNum)
	}

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
		Message:  fmt.Sprintf("Successfully %s on %s.", actionMsg, device.Name),
		DeviceID: device.ID,
	}, nil
}

// extractSwitchNumber extracts switch number from prompt (e.g., "switch 1" -> 1)
func (s *SwitchSensor) extractSwitchNumber(promptLower string) int {
	// Look for patterns like "switch 1", "switch1", "saklar 1", "lampu 1", etc.
	patterns := []string{"switch ", "switch_", "saklar ", "lampu "}
	for _, pattern := range patterns {
		if idx := strings.Index(promptLower, pattern); idx != -1 {
			remaining := promptLower[idx+len(pattern):]
			// Handle direct number after pattern (e.g., "switch1" or "switch_1")
			if len(remaining) > 0 {
				// Skip underscore if present
				if remaining[0] == '_' && len(remaining) > 1 {
					remaining = remaining[1:]
				}
				// Extract first word/number
				words := strings.FieldsFunc(remaining, func(r rune) bool {
					return r == ' ' || r == '_' || r == '-'
				})
				if len(words) > 0 {
					if num, err := strconv.Atoi(words[0]); err == nil && num >= 1 && num <= 10 {
						return num
					}
				}
			}
		}
	}
	return 0
}
