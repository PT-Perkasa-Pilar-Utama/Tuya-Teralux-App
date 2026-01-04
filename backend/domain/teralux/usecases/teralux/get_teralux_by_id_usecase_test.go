package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetTeraluxByID_UserBehavior(t *testing.T) {
	repo, devRepo := setupTestEnv(t) // Need devRepo for associating devices
	useCase := NewGetTeraluxByIDUseCase(repo, devRepo)

	// Seed data
	repo.Create(&entities.Teralux{ID: "t1", Name: "Living Room", MacAddress: "M1", RoomID: "r1"})
	// Seed associated device
	devRepo.Create(&entities.Device{ID: "d1", TeraluxID: "t1", Name: "Light 1"})

	// 1. Get Teralux By ID (Success)
	// URL: GET /api/teralux/t1
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "data": { "id": "t1", "devices": [...] } }
	t.Run("Get Teralux By ID (Success)", func(t *testing.T) {
		res, err := useCase.Execute("t1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.ID != "t1" {
			t.Errorf("Expected ID 't1', got '%s'", res.ID)
		}
		if len(res.Devices) != 1 {
			t.Errorf("Expected 1 device, got %d", len(res.Devices))
		}
		if res.Devices[0].ID != "d1" {
			t.Errorf("Expected Device ID 'd1', got '%s'", res.Devices[0].ID)
		}
	})

	// 2. Get Teralux By ID (Not Found)
	// URL: GET /api/teralux/unknown
	// METHOD: GET
	// RES: 404 Not Found
	// RESPONSE: { "status": false, "message": "record not found", "data": nil }
	t.Run("Get Teralux By ID (Not Found)", func(t *testing.T) {
		_, err := useCase.Execute("unknown")
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") { // GORM standard error
			t.Errorf("Expected 'record not found', got: %v", err)
		}
	})
}
