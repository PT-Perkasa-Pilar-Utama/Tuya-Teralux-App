package usecases

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTuyaGetAllDevicesUseCase_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.ReleaseMode)

	// Setup Config
	utils.AppConfig = &utils.Config{
		TuyaBaseURL:  "https://mock-url",
		TuyaClientID: "client-id",
	}

	deviceService := services.NewTuyaDeviceService()
	// Pass nil for DeviceStateUseCase
	useCase := NewTuyaGetAllDevicesUseCase(deviceService, nil)

	// Execute
	resp, err := useCase.GetAllDevices("token", "uid", 0, 0, "")

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Verify (Mock returns empty list in service)
	if resp.TotalDevices != 0 {
		t.Errorf("Expected 0 devices, got %d", resp.TotalDevices)
	}
}
