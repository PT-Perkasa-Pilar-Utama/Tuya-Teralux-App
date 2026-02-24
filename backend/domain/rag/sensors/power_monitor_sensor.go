package sensors

import (
	"fmt"
	"strings"
	"teralux_app/domain/rag/dtos"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

type PowerMonitorSensor struct{}

func NewPowerMonitorSensor() DeviceSensor {
	return &PowerMonitorSensor{}
}

func (s *PowerMonitorSensor) CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool {
	if device.Category == "dlq" {
		return true
	}

	for _, status := range device.Status {
		if status.Code == "cur_power" || status.Code == "cur_current" ||
			status.Code == "cur_voltage" || status.Code == "add_ele" {
			return true
		}
	}

	return false
}

func (s *PowerMonitorSensor) ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, executor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error) {
	promptLower := strings.ToLower(prompt)

	isStatusQuery := strings.Contains(promptLower, "status") || strings.Contains(promptLower, "berapa") ||
		strings.Contains(promptLower, "power") || strings.Contains(promptLower, "daya") ||
		strings.Contains(promptLower, "voltage") || strings.Contains(promptLower, "tegangan") ||
		strings.Contains(promptLower, "current") || strings.Contains(promptLower, "arus")

	if isStatusQuery {
		return s.getMonitoringStatus(device)
	}

	var commands []tuyaDtos.TuyaCommandDTO
	var actionMsg string

	isOn := strings.Contains(promptLower, "on") || strings.Contains(promptLower, "nyalakan") || strings.Contains(promptLower, "hidupkan")
	isOff := strings.Contains(promptLower, "off") || strings.Contains(promptLower, "matikan") || strings.Contains(promptLower, "mati")

	if !isOn && !isOff {
		return s.getMonitoringStatus(device)
	}

	var switchCode string
	for _, status := range device.Status {
		if status.Code == "switch" || strings.HasPrefix(status.Code, "switch_") {
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

func (s *PowerMonitorSensor) getMonitoringStatus(device *tuyaDtos.TuyaDeviceDTO) (*dtos.ControlResultDTO, error) {
	var power float64
	var current float64
	var voltage float64
	var energy float64
	var temperature int = -1
	var switchState string

	for _, status := range device.Status {
		switch status.Code {
		case "cur_power":
			if val, ok := status.Value.(float64); ok {
				power = val / 10.0
			}
		case "cur_current":
			if val, ok := status.Value.(float64); ok {
				current = val / 1000.0
			}
		case "cur_voltage":
			if val, ok := status.Value.(float64); ok {
				voltage = val / 10.0
			}
		case "add_ele":
			if val, ok := status.Value.(float64); ok {
				energy = val / 100.0
			}
		case "temp_value":
			if val, ok := status.Value.(float64); ok {
				temperature = int(val)
			}
		case "switch":
			if val, ok := status.Value.(bool); ok {
				if val {
					switchState = "ON"
				} else {
					switchState = "OFF"
				}
			}
		}
	}

	message := fmt.Sprintf("⚡ %s Monitoring:\n", device.Name)

	if switchState != "" {
		icon := "🔴"
		if switchState == "ON" {
			icon = "🟢"
		}
		message += fmt.Sprintf("%s Status: %s\n", icon, switchState)
	}

	if power > 0 {
		message += fmt.Sprintf("💡 Power: %.1f W\n", power)
	}

	if voltage > 0 {
		message += fmt.Sprintf("⚡ Voltage: %.1f V\n", voltage)
	}

	if current > 0 {
		message += fmt.Sprintf("🔌 Current: %.3f A\n", current)
	}

	if energy > 0 {
		message += fmt.Sprintf("📊 Energy: %.2f kWh\n", energy)
	}

	if temperature > 0 {
		tempIcon := "🌡️"
		if temperature > 50 {
			tempIcon = "🔥"
		}
		message += fmt.Sprintf("%s Temperature: %d°C", tempIcon, temperature)
	}

	return &dtos.ControlResultDTO{
		Message:  message,
		DeviceID: device.ID,
	}, nil
}
