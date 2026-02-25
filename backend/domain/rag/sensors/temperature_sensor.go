package sensors

import (
	"fmt"
	"sensio/domain/rag/dtos"
	tuyaDtos "sensio/domain/tuya/dtos"
	tuyaUsecases "sensio/domain/tuya/usecases"
)

type TemperatureSensor struct{}

func NewTemperatureSensor() DeviceSensor {
	return &TemperatureSensor{}
}

func (s *TemperatureSensor) CanHandle(device *tuyaDtos.TuyaDeviceDTO) bool {
	// Handle temperature and humidity sensors (wsdcg category)
	if device.Category == "wsdcg" {
		return true
	}

	// Also handle devices with temperature/humidity status
	for _, status := range device.Status {
		if status.Code == "va_temperature" || status.Code == "va_humidity" ||
			status.Code == "temp_current" || status.Code == "humidity_value" {
			return true
		}
	}

	return false
}

func (s *TemperatureSensor) ExecuteControl(token string, device *tuyaDtos.TuyaDeviceDTO, prompt string, history []string, executor tuyaUsecases.TuyaDeviceControlExecutor) (*dtos.ControlResultDTO, error) {
	// Temperature sensors are typically read-only, return current readings
	var temperature float64
	var humidity int
	var battery int = -1

	for _, status := range device.Status {
		switch status.Code {
		case "va_temperature":
			// Temperature is typically in 0.1°C units (e.g., 262 = 26.2°C)
			if val, ok := status.Value.(float64); ok {
				temperature = val / 10.0
			}
		case "temp_current":
			if val, ok := status.Value.(float64); ok {
				temperature = val / 10.0
			}
		case "va_humidity":
			if val, ok := status.Value.(float64); ok {
				humidity = int(val)
			}
		case "humidity_value":
			if val, ok := status.Value.(float64); ok {
				humidity = int(val)
			}
		case "battery_percentage":
			if val, ok := status.Value.(float64); ok {
				battery = int(val)
			}
		}
	}

	// Build response message
	message := fmt.Sprintf("📊 %s Status:\n", device.Name)

	if temperature > 0 {
		message += fmt.Sprintf("🌡️ Temperature: %.1f°C\n", temperature)
	}

	if humidity > 0 {
		message += fmt.Sprintf("💧 Humidity: %d%%\n", humidity)
	}

	if battery >= 0 {
		batteryIcon := "🔋"
		if battery < 20 {
			batteryIcon = "🪫"
		}
		message += fmt.Sprintf("%s Battery: %d%%", batteryIcon, battery)
	}

	return &dtos.ControlResultDTO{
		Message:  message,
		DeviceID: device.ID,
	}, nil
}
