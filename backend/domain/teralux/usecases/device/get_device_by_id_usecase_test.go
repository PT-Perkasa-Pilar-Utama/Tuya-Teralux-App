package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDeviceByIDUseCase_UserBehavior(t *testing.T) {
	repo, _, _ := setupDeviceTestEnv(t)
	useCase := NewGetDeviceByIDUseCase(repo)

	repo.Create(&entities.Device{ID: "dev-1", Name: "Kitchen Switch", TeraluxID: "tx-1"})

	// 1. Get Device By ID (Success)
	// URL: GET /api/devices/dev-1
	// SCENARIO: Device exists.
	// RES: 200 OK
	t.Run("Get Device By ID (Success)", func(t *testing.T) {
		res, err := useCase.Execute("dev-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Device.TeraluxID != "tx-1" {
			t.Errorf("Expected TeraluxID 'tx-1', got '%s'", res.Device.TeraluxID)
		}
		if res.Device.Name != "Kitchen Switch" {
			t.Errorf("Expected Name 'Kitchen Switch', got '%s'", res.Device.Name)
		}
	})

	// 2. Get Device By ID (Not Found)
	// URL: GET /api/devices/dev-unknown
	// SCENARIO: Device does not exist.
	// RES: 404 Not Found
	t.Run("Get Device By ID (Not Found)", func(t *testing.T) {
		_, err := useCase.Execute("dev-unknown")
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 3. Security: Unauthorized
}
