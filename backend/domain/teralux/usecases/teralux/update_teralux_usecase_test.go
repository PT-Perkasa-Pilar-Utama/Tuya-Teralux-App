package usecases

import (
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestUpdateTeralux_UserBehavior(t *testing.T) {
	repo, _ := setupTestEnv(t)
	useCase := NewUpdateTeraluxUseCase(repo)

	// Seed data
	repo.Create(&entities.Teralux{ID: "t1", Name: "Old Name", MacAddress: "AA:BB", RoomID: "r1"})

	// 1. Update Teralux (Success - Partial Update)
	// URL: PUT /api/teralux/t1
	// METHOD: PUT
	// BODY: { "name": "New Name" }
	// RES: 200 OK
	// RESPONSE: { "status": true, "message": "Updated successfully" }
	t.Run("Update Teralux (Success - Partial Update)", func(t *testing.T) {
		req := &dtos.UpdateTeraluxRequestDTO{
			Name: "New Name",
		}

		err := useCase.Execute("t1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify update using repo
		updated, _ := repo.GetByID("t1")
		if updated.Name != "New Name" {
			t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
		}
		if updated.MacAddress != "AA:BB" {
			t.Error("MacAddress should not change if not provided")
		}
	})

	// 2. Update Teralux (Not Found)
	// URL: PUT /api/teralux/unknown
	// METHOD: PUT
	// BODY: { "name": "Hack" }
	// RES: 404 Not Found
	// RESPONSE: { "status": false, "message": "record not found" }
	t.Run("Update Teralux (Not Found)", func(t *testing.T) {
		req := &dtos.UpdateTeraluxRequestDTO{Name: "Hack"}
		err := useCase.Execute("unknown", req)
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'record not found', got: %v", err)
		}
	})
}
