package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDevicesByTeraluxIDUseCase_UserBehavior(t *testing.T) {
	repo, _, _ := setupDeviceTestEnv(t)
	useCase := NewGetDevicesByTeraluxIDUseCase(repo)

	// Seed data
	repo.Create(&entities.Device{ID: "d1", Name: "Light 1", TeraluxID: "tx-1"})
	repo.Create(&entities.Device{ID: "d2", Name: "Light 2", TeraluxID: "tx-1"})
	repo.Create(&entities.Device{ID: "d3", Name: "Fan", TeraluxID: "tx-2"})

	// 1. Get Devices By Teralux ID (Success)
	// URL: GET /api/devices/teralux/tx-1
	// SCENARIO: Retrieve all devices linked to a specific Teralux ID.
	// RES: 200 OK
	t.Run("Get Devices By Teralux ID (Success)", func(t *testing.T) {
		res, err := useCase.Execute("tx-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Total != 2 {
			t.Errorf("Expected 2 devices for tx-1, got %d", res.Total)
		}
		for _, d := range res.Devices {
			if d.TeraluxID != "tx-1" {
				t.Errorf("Expected TeraluxID 'tx-1', got '%s'", d.TeraluxID)
			}
		}
	})

	// 2. Get Devices By Teralux ID (Success - Empty)
	// URL: GET /api/devices/teralux/tx-empty
	// SCENARIO: No devices exist for this Teralux ID.
	// RES: 200 OK
	t.Run("Get Devices By Teralux ID (Success - Empty)", func(t *testing.T) {
		res, err := useCase.Execute("tx-empty")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Total != 0 {
			t.Errorf("Expected 0 devices, got %d", res.Total)
		}
		if len(res.Devices) != 0 {
			t.Errorf("Expected empty list, got %d", len(res.Devices))
		}
	})
}
