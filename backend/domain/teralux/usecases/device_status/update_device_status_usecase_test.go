package usecases

import (
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestUpdateDeviceStatus_UserBehavior(t *testing.T) {
	repo := setupStatusTestEnv(t)
	useCase := NewUpdateDeviceStatusUseCase(repo)

	// Seed data
	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_1", Value: "false"})

	// 1. Update Device Status (Success)
	// URL: PUT /api/devices/statuses/dev-1/switch_1
	// METHOD: PUT
	// BODY: { "value": "true" }
	// RES: 200 OK
	// RESPONSE: { "status": true, "message": "Device status updated successfully" }
	t.Run("Update Device Status (Success)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Value: "true"}
		err := useCase.Execute("dev-1", "switch_1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify update in repo
		updated, _ := repo.GetByDeviceIDAndCode("dev-1", "switch_1")
		if updated.Value != "true" {
			t.Errorf("Expected value 'true', got '%s'", updated.Value)
		}
	})

	// 2. Update Device Status (Not Found)
	// URL: PUT /api/devices/statuses/unknown-dev/switch_1
	// METHOD: PUT
	// BODY: { "value": "true" }
	// RES: 404 Not Found
	// RESPONSE: { "status": false, "message": "record not found" }
	t.Run("Update Device Status (Not Found)", func(t *testing.T) {
		req := &dtos.UpdateDeviceStatusRequestDTO{Value: "true"}
		err := useCase.Execute("unknown-dev", "switch_1", req)
		if err == nil {
			t.Fatal("Expected error for unknown status, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'record not found', got: %v", err)
		}
	})

	// 3. Update Device Status (No Change - Empty Value)
	// URL: PUT /api/devices/statuses/dev-1/switch_1
	// METHOD: PUT
	// BODY: { "value": "" }
	// RES: 200 OK
	// RESPONSE: { "status": true, "message": "Device status updated successfully" }
	t.Run("Update Device Status (No Change - Empty Value)", func(t *testing.T) {
		// Reset state
		repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_1", Value: "false"})

		req := &dtos.UpdateDeviceStatusRequestDTO{Value: ""}
		err := useCase.Execute("dev-1", "switch_1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify value unchanged
		updated, _ := repo.GetByDeviceIDAndCode("dev-1", "switch_1")
		if updated.Value != "false" {
			t.Errorf("Expected value 'false' (unchanged), got '%s'", updated.Value)
		}
	})
}
