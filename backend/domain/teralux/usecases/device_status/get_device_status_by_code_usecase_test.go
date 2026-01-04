package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDeviceStatusByCode_UserBehavior(t *testing.T) {
	repo := setupStatusTestEnv(t)
	useCase := NewGetDeviceStatusByCodeUseCase(repo)

	// Seed data
	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_1", Value: "true"})

	// 1. Get Device Status (Success)
	// URL: GET /api/devices/statuses/dev-1/switch_1
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "data": { "device_id": "dev-1", "code": "switch_1", "value": "true" } }
	t.Run("Get Device Status (Success)", func(t *testing.T) {
		res, err := useCase.Execute("dev-1", "switch_1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Value != "true" {
			t.Errorf("Expected value 'true', got '%s'", res.Value)
		}
		if res.Code != "switch_1" {
			t.Errorf("Expected code 'switch_1', got '%s'", res.Code)
		}
	})

	// 2. Get Device Status (Not Found - Unknown Device)
	// URL: GET /api/devices/statuses/unknown-dev/switch_1
	// METHOD: GET
	// RES: 404 Not Found
	// RESPONSE: { "status": false, "message": "record not found", "data": nil }
	t.Run("Get Device Status (Not Found - Unknown Device)", func(t *testing.T) {
		_, err := useCase.Execute("unknown-dev", "switch_1")
		if err == nil {
			t.Fatal("Expected error for unknown device, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'record not found', got: %v", err)
		}
	})

	// 3. Get Device Status (Not Found - Unknown Code)
	// URL: GET /api/devices/statuses/dev-1/unknown-code
	// METHOD: GET
	// RES: 404 Not Found
	// RESPONSE: { "status": false, "message": "record not found", "data": nil }
	t.Run("Get Device Status (Not Found - Unknown Code)", func(t *testing.T) {
		_, err := useCase.Execute("dev-1", "unknown-code")
		if err == nil {
			t.Fatal("Expected error for unknown code, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'record not found', got: %v", err)
		}
	})
}
