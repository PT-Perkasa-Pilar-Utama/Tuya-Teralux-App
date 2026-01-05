package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDeviceStatusByCode_UserBehavior(t *testing.T) {
	repo, devRepo := setupStatusTestEnv(t)
	useCase := NewGetDeviceStatusByCodeUseCase(repo, devRepo)

	// Seed data
	devRepo.Create(&entities.Device{ID: "dev-1", Name: "D1"})
	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_1", Value: "true"})

	// 1. Get Status By Code (Success)
	// URL: GET /api/device-statuses/code/switch_1?device_id=dev-1
	// RES: 200 OK
	t.Run("Get Status By Code (Success)", func(t *testing.T) {
		res, err := useCase.Execute("dev-1", "switch_1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.DeviceStatus.Value != "true" {
			t.Errorf("Expected value 'true', got '%s'", res.DeviceStatus.Value)
		}
	})

	// 2. Get Status By Code (Not Found - Code)
	// URL: GET /api/device-statuses/code/unknown_code?device_id=dev-1
	// RES: 404 Not Found
	t.Run("Get Status By Code (Not Found - Code)", func(t *testing.T) {
		_, err := useCase.Execute("dev-1", "unknown_code")
		if err == nil {
			t.Fatal("Expected error for unknown code, got nil")
		}
		if !strings.Contains(err.Error(), "Status code not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'Status code not found', got: %v", err)
		}
	})

	// 3. Get Status By Code (Not Found - Device)
	// URL: GET /api/device-statuses/code/switch_1?device_id=unknown
	// RES: 404 Not Found
	t.Run("Get Status By Code (Not Found - Device)", func(t *testing.T) {
		_, err := useCase.Execute("unknown", "switch_1")
		if err == nil {
			t.Fatal("Expected error for unknown device, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'Device not found', got: %v", err)
		}
	})

	// 4. Validation: Missing Device ID
	// URL: GET /api/device-statuses/code/switch_1
	// RES: 400 Bad Request
	t.Run("Validation: Missing Device ID", func(t *testing.T) {
		_, err := useCase.Execute("", "switch_1")
		if err == nil {
			t.Fatal("Expected error for missing device ID, got nil")
		}
		if !strings.Contains(err.Error(), "device_id is required") {
			t.Errorf("Expected 'device_id is required', got: %v", err)
		}
	})

	// 5. Unauthorized (Middleware)
}
