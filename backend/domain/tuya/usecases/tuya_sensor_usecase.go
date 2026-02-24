package usecases

import (
	"fmt"
	"teralux_app/domain/tuya/dtos"
)

// TuyaSensorUseCase handles retrieval and interpretation of sensor data.
// It parses raw device status values (like temperature, humidity) into formatted DTOs.
type TuyaSensorUseCase struct {
	getDeviceUseCase *TuyaGetDeviceByIDUseCase
}

// NewTuyaSensorUseCase initializes a new TuyaSensorUseCase.
//
// param getDeviceUseCase The usecase dependency for fetching raw device data.
// return *TuyaSensorUseCase A pointer to the initialized usecase.
func NewTuyaSensorUseCase(getDeviceUseCase *TuyaGetDeviceByIDUseCase) *TuyaSensorUseCase {
	return &TuyaSensorUseCase{
		getDeviceUseCase: getDeviceUseCase,
	}
}

// GetSensorData retrieves, interprets, and formats sensor readings for a specific device.
// It converts raw values (often integers scaled by 10) into human-readable floats and generates descriptive status text.
//
// param accessToken The valid OAuth 2.0 access token.
// param deviceID The device ID of the sensor.
// return *dtos.SensorDataDTO The structured sensor data containing temperature, humidity, and status.
// return error An error if fetching the device data fails.
func (uc *TuyaSensorUseCase) GetSensorData(accessToken, deviceID string) (*dtos.SensorDataDTO, error) {
	device, err := uc.getDeviceUseCase.GetDeviceByID(accessToken, deviceID)
	if err != nil {
		return nil, err
	}

	var temperature float64
	var humidity int
	var battery int

	// Parse status values
	for _, status := range device.Status {
		switch status.Code {
		case "va_temperature":
			// value is likely float64 or int in JSON, often comes as float64 from generic interface{} unmarshaling
			switch val := status.Value.(type) {
			case float64:
				temperature = val / 10.0
			case int:
				temperature = float64(val) / 10.0
			}
		case "va_humidity":
			if val, ok := status.Value.(float64); ok {
				humidity = int(val)
			}
		case "battery_percentage":
			if val, ok := status.Value.(float64); ok {
				battery = int(val)
			}
		}
	}

	// Determine status text
	var tempStatus string
	switch {
	case temperature > 28.0:
		tempStatus = "Temperature hot"
	case temperature < 18.0:
		tempStatus = "Temperature cold"
	default:
		tempStatus = "Temperature comfortable"
	}

	var humidStatus string
	switch {
	case humidity > 60:
		humidStatus = "Air moist"
	case humidity < 30:
		humidStatus = "Air dry"
	default:
		humidStatus = "Air comfortable"
	}

	statusText := fmt.Sprintf("%s, %s", tempStatus, humidStatus)

	response := &dtos.SensorDataDTO{
		Temperature:       temperature,
		Humidity:          humidity,
		BatteryPercentage: battery,
		StatusText:        statusText,
		TempUnit:          "°C", // Defaulting as per plan
	}

	return response, nil
}
