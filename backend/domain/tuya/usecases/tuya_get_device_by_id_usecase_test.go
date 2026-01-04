package usecases

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTuyaGetDeviceByIDUseCase_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.ReleaseMode)

	utils.AppConfig = &utils.Config{
		TuyaBaseURL:  "https://mock-url",
		TuyaClientID: "client-id",
	}

	deviceService := services.NewTuyaDeviceService()
	// NewTuyaGetDeviceByIDUseCase(service, deviceStateUC)

	useCase := NewTuyaGetDeviceByIDUseCase(deviceService, nil)

	// Execute
	resp, err := useCase.GetDeviceByID("test-token", "mock-device-id")

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if resp.ID != "mock-device-id" {
		t.Errorf("Expected mock-device-id, got %s", resp.ID)
	}
}

func TestTuyaGetDeviceByIDUseCase_Execute_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.ReleaseMode)

	utils.AppConfig = &utils.Config{}

	deviceService := services.NewTuyaDeviceService()
	useCase := NewTuyaGetDeviceByIDUseCase(deviceService, nil)

	// Invoke with logic that might fail or mocked logic
	_, err := useCase.GetDeviceByID("token", "invalid_device_id_99999")
	if err == nil {
		t.Error("Expected error for invalid device id")
	}
}
