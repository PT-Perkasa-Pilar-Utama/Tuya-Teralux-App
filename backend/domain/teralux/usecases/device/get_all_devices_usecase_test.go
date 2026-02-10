package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetAllDevicesUseCase_UserBehavior(t *testing.T) {
	repo, _, _ := setupDeviceTestEnv(t)
	useCase := NewGetAllDevicesUseCase(repo)

	// Seed data
	_ = repo.Create(&entities.Device{ID: "d1", Name: "Light 1", TeraluxID: "tx-1"})
	_ = repo.Create(&entities.Device{ID: "d2", Name: "Light 2", TeraluxID: "tx-1"})
	_ = repo.Create(&entities.Device{ID: "d3", Name: "Fan", TeraluxID: "tx-2"})

	// 1. Get All Devices (Success - Filter by Teralux)
	// URL: GET /api/devices?teralux_id=tx-1
	// SCENARIO: Filter by Teralux Hub.
	// RES: 200 OK
	t.Run("Get All Devices (Success - Filter by Teralux)", func(t *testing.T) {
		teraID := "tx-1"
		filter := &dtos.DeviceFilterDTO{TeraluxID: &teraID}
		res, err := useCase.Execute(filter)
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

	// 2. Get All Devices (Success - Empty)
	// URL: GET /api/devices?teralux_id=tx-999
	// SCENARIO: No devices match filter (or system empty).
	// RES: 200 OK
	t.Run("Get All Devices (Success - Empty)", func(t *testing.T) {
		teraID := "tx-999"
		filter := &dtos.DeviceFilterDTO{TeraluxID: &teraID}
		res, err := useCase.Execute(filter)
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

	// 3. Security: Unauthorized
}
