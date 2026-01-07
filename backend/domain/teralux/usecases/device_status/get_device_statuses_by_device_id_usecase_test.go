package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDeviceStatusesByDeviceID_UserBehavior(t *testing.T) {
	repo, devRepo := setupStatusTestEnv(t)
	// We assume New... takes both repos now
	useCase := NewGetDeviceStatusesByDeviceIDUseCase(repo, devRepo)

	// Seed data
	devRepo.Create(&entities.Device{ID: "dev-1", Name: "D1"})
	devRepo.Create(&entities.Device{ID: "dev-2", Name: "D2"}) // Empty statuses

	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_1", Value: "true"})
	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_2", Value: "false"})

	// 1. Get Statuses By Device ID (Success)
	// URL: GET /api/devices/dev-1/statuses
	// RES: 200 OK
	t.Run("Get Statuses By Device ID (Success)", func(t *testing.T) {
		res, err := useCase.Execute("dev-1", 0, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res.DeviceStatuses) != 2 {
			t.Errorf("Expected 2 statuses, got %d", len(res.DeviceStatuses))
		}
	})

	// 2. Get Statuses By Device ID (Success - Empty)
	// URL: GET /api/devices/dev-2/statuses
	// RES: 200 OK
	t.Run("Get Statuses By Device ID (Success - Empty)", func(t *testing.T) {
		res, err := useCase.Execute("dev-2", 0, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res.DeviceStatuses) != 0 {
			t.Errorf("Expected empty list, got %d items", len(res.DeviceStatuses))
		}
	})

	// 3. Get Statuses By Device ID (Not Found)
	// URL: GET /api/devices/unknown/statuses
	// RES: 404 Not Found
	t.Run("Get Statuses By Device ID (Not Found)", func(t *testing.T) {
		_, err := useCase.Execute("unknown", 0, 0)
		if err == nil {
			t.Fatal("Expected error for unknown device, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 4. Unauthorized (Middleware)
}
