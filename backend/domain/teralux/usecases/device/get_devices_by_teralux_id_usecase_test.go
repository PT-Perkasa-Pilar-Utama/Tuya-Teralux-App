package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDevicesByTeraluxIDUseCase_UserBehavior(t *testing.T) {
	repo, _, teraluxRepo := setupDeviceTestEnv(t)
	useCase := NewGetDevicesByTeraluxIDUseCase(repo, teraluxRepo)

	// Seed data
	teraluxRepo.Create(&entities.Teralux{ID: "tx-1", Name: "Hub 1", MacAddress: "M1", RoomID: "r1"})
	teraluxRepo.Create(&entities.Teralux{ID: "tx-empty", Name: "Hub Empty", MacAddress: "M2", RoomID: "r1"})

	repo.Create(&entities.Device{ID: "d1", Name: "Light 1", TeraluxID: "tx-1"})
	repo.Create(&entities.Device{ID: "d2", Name: "Light 2", TeraluxID: "tx-1"})
	repo.Create(&entities.Device{ID: "d3", Name: "Fan", TeraluxID: "tx-2"})

	// 1. Get Devices By Teralux ID (Success)
	// URL: GET /api/devices/teralux/tx-1
	// SCENARIO: Retrieve all devices linked to a specific Teralux ID.
	// RES: 200 OK
	// 1. Get Devices By Teralux ID (Success)
	// URL: GET /api/devices/teralux/tx-1
	// SCENARIO: Retrieve all devices linked to a specific Teralux ID.
	// RES: 200 OK
	t.Run("Get Devices By Teralux ID (Success)", func(t *testing.T) {
		res, err := useCase.Execute("tx-1", 0, 0)
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
		res, err := useCase.Execute("tx-empty", 0, 0)
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

	// 3. Get Devices By Teralux ID (Not Found - Teralux ID)
	// URL: GET /api/devices/teralux/tx-999
	// SCENARIO: Teralux ID does not exist.
	// RES: 404 Not Found
	t.Run("Get Devices By Teralux ID (Not Found - Teralux ID)", func(t *testing.T) {
		_, err := useCase.Execute("tx-999", 0, 0)
		if err == nil {
			t.Fatal("Expected error for unknown Teralux ID, got nil")
		}
		if err.Error() != "Teralux hub not found: record not found" {
			t.Errorf("Expected 'Teralux hub not found' error, got: %v", err)
		}
	})
}
