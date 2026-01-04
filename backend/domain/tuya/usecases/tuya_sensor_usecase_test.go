package usecases

import (
	"net/http"
	"net/http/httptest"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTuyaSensorUseCase_GetSensorData(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	// Mock Response with status
	mockResp := `{
		"success": true,
		"result": {
			"id": "sensor-1",
			"category": "wsdcg",
			"status": [
				{"code": "va_temperature", "value": 255},
				{"code": "va_humidity", "value": 60},
				{"code": "battery_percentage", "value": 95}
			]
		}
	}`
	// Note: va_temperature 255 -> 25.5

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResp))
	}))
	defer server.Close()

	utils.AppConfig = &utils.Config{
		TuyaBaseURL:  server.URL,
		TuyaClientID: "id",
	}

	devService := services.NewTuyaDeviceService()
	// Pass nil for deviceStateUC as we test sensor logic, not state merging
	getDeviceUC := NewTuyaGetDeviceByIDUseCase(devService, nil)
	sensorUC := NewTuyaSensorUseCase(getDeviceUC)

	// Execute
	data, err := sensorUC.GetSensorData("token", "sensor-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify Logic
	if data.Temperature != 25.5 {
		t.Errorf("Expected temp 25.5, got %v", data.Temperature)
	}
	if data.Humidity != 60 {
		t.Errorf("Expected humid 60, got %v", data.Humidity)
	}
	if data.BatteryPercentage != 95 {
		t.Errorf("Expected battery 95, got %v", data.BatteryPercentage)
	}
	if data.TempUnit != "°C" {
		t.Errorf("Expected unit °C, got %v", data.TempUnit)
	}
}
