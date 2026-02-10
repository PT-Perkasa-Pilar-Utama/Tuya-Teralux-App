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
	_ = repo.Create(&entities.Teralux{ID: "t1", Name: "Living Room", MacAddress: "M1", RoomID: "r1"})
	// Seed associated device
	_ = devRepo.Create(&entities.Device{ID: "d1", TeraluxID: "t1", Name: "Light 1"})

	// 1. Get Teralux By ID (Success)
	// URL: GET /api/teralux/t1
	// SCENARIO: Device exists.
	// RES: 200 OK
	t.Run("Get Teralux By ID (Success)", func(t *testing.T) {
		res, err := useCase.Execute("t1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Teralux.ID != "t1" {
			t.Errorf("Expected ID 't1', got '%s'", res.Teralux.ID)
		}
		if res.Teralux.Name != "Living Room" {
			t.Errorf("Expected Name 'Living Room', got '%s'", res.Teralux.Name)
		}
	})

	// 2. Get Teralux By ID (Not Found)
	// URL: GET /api/teralux/unknown-id
	// SCENARIO: Device does not exist.
	// RES: 404 Not Found
	t.Run("Get Teralux By ID (Not Found)", func(t *testing.T) {
		_, err := useCase.Execute("unknown-id")
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") && !strings.Contains(err.Error(), "Teralux not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 3. Validation: Invalid ID Format
	// URL: GET /api/teralux/INVALID-FORMAT
	// SCENARIO: ID is not a valid UUID format.
	// RES: 400 Bad Request
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		_, err := useCase.Execute("INVALID-FORMAT")
		if err == nil {
			t.Fatal("Expected error for invalid ID format, got nil")
		}
		if err.Error() != "Invalid ID format" {
			t.Errorf("Expected 'Invalid ID format', got: %v", err)
		}
	})

	// 4. Security: Unauthorized
}
