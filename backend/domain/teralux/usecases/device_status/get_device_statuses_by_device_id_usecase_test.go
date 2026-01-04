package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDeviceStatusesByDeviceID_UserBehavior(t *testing.T) {
	repo := setupStatusTestEnv(t)
	useCase := NewGetDeviceStatusesByDeviceIDUseCase(repo)

	// Seed data
	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_1", Value: "true"})
	repo.Upsert(&entities.DeviceStatus{DeviceID: "dev-1", Code: "switch_2", Value: "false"})

	// 1. Get Device Statuses (Success - With Data)
	// URL: GET /api/devices/statuses/dev-1
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "data": [ { "code": "switch_1", ... }, ... ] }
	t.Run("Get Device Statuses (Success - With Data)", func(t *testing.T) {
		res, err := useCase.Execute("dev-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res) != 2 {
			t.Errorf("Expected 2 statuses, got %d", len(res))
		}
	})

	// 2. Get Device Statuses (Success - Empty List / Unknown Device)
	// URL: GET /api/devices/statuses/unknown-dev
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "data": [] }
	// Note: Currently returns empty list if no statuses found, even if device doesn't exist.
	t.Run("Get Device Statuses (Success - Empty List)", func(t *testing.T) {
		res, err := useCase.Execute("unknown-dev")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("Expected empty list, got %d items", len(res))
		}
	})
}
