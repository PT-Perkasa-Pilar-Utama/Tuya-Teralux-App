package usecases

import (
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestUpdateDeviceStatus_UserBehavior(t *testing.T) {
	repo, devRepo := setupStatusTestEnv(t)
	// NewUpdateDeviceStatusUseCase might need DeviceRepo to verify device existence
	// And potentially logic to check valid codes/values.
	// We'll pass DeviceRepo.
	useCase := NewUpdateDeviceStatusUseCase(repo, devRepo)

	// Seed data
	devRepo.Create(&entities.Device{ID: "d1", Name: "D1"})
	repo.Upsert(&entities.DeviceStatus{DeviceID: "d1", Code: "switch_1", Value: "false"})
	repo.Upsert(&entities.DeviceStatus{DeviceID: "d1", Code: "dimmer", Value: "50"})

	// 1. Update Status (Success)
	// URL: PUT /api/devices/d1/status
	// BODY: { "code": "switch_1", "value": true }
	// RES: 200 OK
	t.Run("Update Status (Success)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}
		err := useCase.Execute("d1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		updated, _ := repo.GetByDeviceIDAndCode("d1", "switch_1")
		if updated.Value != "true" { // Assuming stored as string "true"
			t.Errorf("Expected value 'true', got '%s'", updated.Value)
		}
	})

	// 2. Update Status (Not Found - Device)
	// URL: PUT /api/devices/unknown/status
	// RES: 404 Not Found
	t.Run("Update Status (Not Found - Device)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}
		err := useCase.Execute("unknown", req)
		if err == nil {
			t.Fatal("Expected error for unknown device, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'Device not found', got: %v", err)
		}
	})

	// 3. Update Status (Not Found - Invalid Code)
	// URL: PUT /api/devices/d1/status
	// BODY: { "code": "nuclear_launch", "value": true }
	// RES: 404 Not Found
	t.Run("Update Status (Not Found - Invalid Code)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "nuclear_launch", Value: true}
		err := useCase.Execute("d1", req)
		if err == nil {
			t.Fatal("Expected error for invalid code, got nil")
		}
		if !strings.Contains(err.Error(), "Invalid status code") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'Invalid status code', got: %v", err)
		}
	})

	// 4. Validation: Invalid Value Type
	// URL: PUT /api/devices/d1/status
	// BODY: { "code": "dimmer", "value": "full_power" }
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Invalid Value Type", func(t *testing.T) {
		// If DTO `Value` allows "full_power" string, validation must check it against code "dimmer".
		// We'll pass "full_power" string.
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "dimmer", Value: "full_power"}
		err := useCase.Execute("d1", req)
		if err == nil {
			t.Fatal("Expected error for invalid value type, got nil")
		}
		// Scenario says: "Invalid value for status code 'dimmer'"
		// utils.ValidationError
		if !strings.Contains(err.Error(), "Invalid value") {
			t.Errorf("Expected 'Invalid value', got: %v", err)
		}
	})

	// 5. Unauthorized (Middleware)
}
