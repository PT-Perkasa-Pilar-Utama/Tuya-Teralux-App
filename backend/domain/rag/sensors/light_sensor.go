package sensors

import (
	"fmt"
	"strconv"
	"strings"
	"teralux_app/domain/rag/dtos"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type LightSensor struct{}

func NewLightSensor() DeviceSensor {
	return &LightSensor{}
}

func (s *LightSensor) CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool {
	// Handle LED light devices (dj category)
	if device.Category == "dj" {
		return true
	}

	// Also handle devices with LED light status codes
	for _, status := range device.Status {
		if status.Code == "switch_led" || status.Code == "bright_value_v2" ||
			status.Code == "work_mode" || status.Code == "colour_data_v2" {
			return true
		}
	}

	return false
}

func (s *LightSensor) ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, executor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error) {
	promptLower := strings.ToLower(prompt)

	// Check for specific control types
	if code, value, matched := s.matchLightControl(promptLower, device.Status); matched {
		commands := []tuyaDtos.TuyaCommandDTO{{
			Code:  code,
			Value: value,
		}}

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
			Message:  fmt.Sprintf("Successfully controlled %s on %s.", code, device.Name),
			DeviceID: device.ID,
		}, nil
	}

	// Default switch control
	var commands []tuyaDtos.TuyaCommandDTO
	var actionMsg string

	isOn := strings.Contains(promptLower, "on") || strings.Contains(promptLower, "nyalakan") || strings.Contains(promptLower, "hidupkan")
	isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "matikan") || strings.Contains(promptLower, "mati")

	if !isOn && !isOff {
		isOn = true // Default to ON
	}

	// Find switch_led
	var switchCode string
	for _, status := range device.Status {
		if status.Code == "switch_led" {
			switchCode = status.Code
			break
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

// matchLightControl tries to match light-specific controls
func (s *LightSensor) matchLightControl(promptLower string, status []tuyaDtos.TuyaDeviceStatusDTO) (string, interface{}, bool) {
	// Brightness controls
	if strings.Contains(promptLower, "brightness") || strings.Contains(promptLower, "terang") ||
		strings.Contains(promptLower, "redup") || strings.Contains(promptLower, "bright") {
		for _, st := range status {
			if st.Code == "bright_value_v2" || st.Code == "bright_value" {
				// Extract brightness value (0-1000 scale typically)
				if val, found := s.extractNumericValue(promptLower, 0, 1000); found {
					return st.Code, val, true
				}
				// Default brightness keywords
				if strings.Contains(promptLower, "max") || strings.Contains(promptLower, "maksimal") {
					return st.Code, 1000, true
				}
				if strings.Contains(promptLower, "min") || strings.Contains(promptLower, "minimal") {
					return st.Code, 100, true
				}
				if strings.Contains(promptLower, "terang") {
					return st.Code, 800, true
				}
				if strings.Contains(promptLower, "redup") {
					return st.Code, 300, true
				}
				return st.Code, 500, true // Medium
			}
		}
	}

	// Work mode controls
	if strings.Contains(promptLower, "mode") || strings.Contains(promptLower, "scene") {
		for _, st := range status {
			if st.Code == "work_mode" {
				// Common modes: white, colour, scene, music
				if strings.Contains(promptLower, "white") || strings.Contains(promptLower, "putih") {
					return st.Code, "white", true
				}
				if strings.Contains(promptLower, "color") || strings.Contains(promptLower, "colour") || strings.Contains(promptLower, "warna") {
					return st.Code, "colour", true
				}
				if strings.Contains(promptLower, "scene") || strings.Contains(promptLower, "sken") {
					return st.Code, "scene", true
				}
				if strings.Contains(promptLower, "music") || strings.Contains(promptLower, "musik") {
					return st.Code, "music", true
				}
			}
		}
	}

	// Temperature value (color temperature)
	if strings.Contains(promptLower, "warm") || strings.Contains(promptLower, "cool") ||
		strings.Contains(promptLower, "hangat") || strings.Contains(promptLower, "dingin") {
		for _, st := range status {
			if st.Code == "temp_value_v2" || st.Code == "temp_value" {
				if strings.Contains(promptLower, "warm") || strings.Contains(promptLower, "hangat") {
					return st.Code, 0, true // Warm white
				}
				if strings.Contains(promptLower, "cool") || strings.Contains(promptLower, "dingin") {
					return st.Code, 1000, true // Cool white
				}
				return st.Code, 500, true // Neutral
			}
		}
	}

	return "", nil, false
}

// extractNumericValue extracts a numeric value from prompt within given range
func (s *LightSensor) extractNumericValue(promptLower string, min, max int) (int, bool) {
	words := strings.Fields(promptLower)
	for _, word := range words {
		if num, err := strconv.Atoi(word); err == nil && num >= min && num <= max {
			return num, true
		}
	}
	return 0, false
}
