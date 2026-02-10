package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockTuyaDeviceControlExecutor mocks the TuyaDeviceControlExecutor interface
type MockTuyaDeviceControlExecutor struct {
	ShouldFail bool
}

func (m *MockTuyaDeviceControlExecutor) SendCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error) {
	if accessToken == "invalid_token_123" {
		return false, fmt.Errorf("mock error: invalid token")
	}
	if m.ShouldFail {
		return false, fmt.Errorf("mock failure")
	}
	return true, nil
}

func (m *MockTuyaDeviceControlExecutor) SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error) {
	if accessToken == "invalid_token_123" {
		return false, fmt.Errorf("mock error: invalid token")
	}
	if m.ShouldFail {
		return false, fmt.Errorf("mock failure")
	}
	return true, nil
}

func TestUpdateDeviceStatus_UserBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo, devRepo := setupStatusTestEnv(t)

	// Setup Mock
	mockTuya := &MockTuyaDeviceControlExecutor{ShouldFail: false}

	useCase := NewUpdateDeviceStatusUseCase(repo, devRepo, mockTuya)

	// Seed data
	_ = devRepo.Create(&entities.Device{ID: "d1", Name: "D1"})
	_ = repo.Upsert(&entities.DeviceStatus{DeviceID: "d1", Code: "switch_1", Value: "false"})
	_ = repo.Upsert(&entities.DeviceStatus{DeviceID: "d1", Code: "dimmer", Value: "50"})

	// 1. Update Status (Success)
	// URL: PUT /api/devices/d1/status
	// BODY: { "code": "switch_1", "value": true }
	// RES: 200 OK
	t.Run("Update Status (Success - Command)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}
		// Use valid token to pass "command" success check in mock service
		err := useCase.Execute("d1", req, "valid_token")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		updated, _ := repo.GetByDeviceIDAndCode("d1", "switch_1")
		if updated.Value != "true" { // Assuming stored as string "true"
			t.Errorf("Expected value 'true', got '%s'", updated.Value)
		}
	})

	// 2. Update Status (Not Found - Device)
	t.Run("Update Status (Not Found - Device)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}
		err := useCase.Execute("unknown", req, "valid_token")
		if err == nil {
			t.Fatal("Expected error for unknown device, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'Device not found', got: %v", err)
		}
	})

	// 3. Update Status (Not Found - Invalid Code)
	t.Run("Update Status (Not Found - Invalid Code)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "nuclear_launch", Value: true}
		err := useCase.Execute("d1", req, "valid_token")
		if err == nil {
			t.Fatal("Expected error for invalid code, got nil")
		}
		if !strings.Contains(err.Error(), "Invalid status code") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'Invalid status code', got: %v", err)
		}
	})

	// 4. Validation: Invalid Value Type
	t.Run("Validation: Invalid Value Type", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "dimmer", Value: "full_power"}
		err := useCase.Execute("d1", req, "valid_token")
		if err == nil {
			t.Fatal("Expected error for invalid value type, got nil")
		}
		if !strings.Contains(err.Error(), "Invalid value") {
			t.Errorf("Expected 'Invalid value', got: %v", err)
		}
	})

	// 5. Command Failure (Invalid Token)
	t.Run("Command Failure (Invalid Token)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}
		err := useCase.Execute("d1", req, "invalid_token_123")
		if err == nil {
			t.Fatal("Expected error for command failure, got nil")
		}
		// Should contain "mock error: invalid token"
		if !strings.Contains(err.Error(), "mock error: invalid token") {
			t.Errorf("Expected 'invalid token', got: %v", err)
		}
	})

	// 6. IR Command Success
	t.Run("IR Command Success", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{
			Code:     "temp",
			Value:    24,
			RemoteID: "ir_remote_1",
		}
		// Reuse d1 or creates new one if needed, but d1 exists.
		// IR logic calls SendIRACCommand.
		// "valid_token" triggers mock success.
		err := useCase.Execute("d1", req, "valid_token")
		if err != nil {
			t.Fatalf("Unexpected error for IR command: %v", err)
		}

		updated, _ := repo.GetByDeviceIDAndCode("d1", "temp")
		if updated.Value != "24" {
			t.Errorf("Expected value '24', got '%s'", updated.Value)
		}
	})
}
