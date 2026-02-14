package usecases

import (
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestUpdateDeviceUseCase_UserBehavior(t *testing.T) {
	repo, _, _ := setupDeviceTestEnv(t)
	useCase := NewUpdateDeviceUseCase(repo)

	_ = repo.Create(&entities.Device{ID: "dev-1", Name: "Old Name"})

	// 1. Update Device (Success)
	// URL: PUT /api/devices/dev-1
	// SCENARIO: Valid payload.
	// RES: 200 OK
	t.Run("Update Device (Success)", func(t *testing.T) {
		newName := "Kitchen Sink Light"
		req := &dtos.UpdateDeviceRequestDTO{Name: &newName}
		err := useCase.UpdateDevice("dev-1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		updated, _ := repo.GetByID("dev-1")
		if updated.Name != "Kitchen Sink Light" {
			t.Errorf("Expected Name 'Kitchen Sink Light', got '%s'", updated.Name)
		}
	})

	// 2. Update Device (Not Found)
	// URL: PUT /api/devices/dev-unknown
	// SCENARIO: Device does not exist.
	// RES: 404 Not Found
	t.Run("Update Device (Not Found)", func(t *testing.T) {
		name := "New Name"
		req := &dtos.UpdateDeviceRequestDTO{Name: &name}
		err := useCase.UpdateDevice("dev-unknown", req)
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 3. Validation: Empty Name
	// URL: PUT /api/devices/dev-1
	// SCENARIO: Name is empty.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Empty Name", func(t *testing.T) {
		emptyName := ""
		req := &dtos.UpdateDeviceRequestDTO{Name: &emptyName}
		err := useCase.UpdateDevice("dev-1", req)
		if err == nil {
			t.Fatal("Expected error for empty name, got nil")
		}
		// Based on scenario: "Validation Error: name cannot be empty"
		// Code should match this error message or use utils.ValidationError
		if !strings.Contains(err.Error(), "name cannot be empty") {
			t.Errorf("Expected 'name cannot be empty', got: %v", err)
		}
	})

	// 4. Security: Unauthorized
}
