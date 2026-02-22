package usecases

import (
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestUpdateTeralux_UserBehavior(t *testing.T) {
	repo, _, _ := setupTestEnv(t)
	useCase := NewUpdateTeraluxUseCase(repo)

	// Seed data
	_ = repo.Create(&entities.Teralux{ID: "t1", Name: "Old Name", MacAddress: "AA:BB", RoomID: "r1"})

	// 1. Update Teralux (Success - Name Only)
	// URL: PUT /api/teralux/t1
	// SCENARIO: Valid payload (name only).
	// RES: 200 OK
	t.Run("Update Teralux (Success - Name Only)", func(t *testing.T) {
		newName := "New Name"
		req := &dtos.UpdateTeraluxRequestDTO{
			Name: &newName,
		}

		err := useCase.UpdateTeralux("t1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		updated, _ := repo.GetByID("t1")
		if updated.Name != "New Name" {
			t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
		}
	})

	// 2. Update Teralux (Success - Move Room)
	// URL: PUT /api/teralux/t1
	// SCENARIO: Valid payload (room_id only).
	// RES: 200 OK
	t.Run("Update Teralux (Success - Move Room)", func(t *testing.T) {
		newRoom := "room-2"
		req := &dtos.UpdateTeraluxRequestDTO{
			RoomID: &newRoom,
		}

		err := useCase.UpdateTeralux("t1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		updated, _ := repo.GetByID("t1")
		if updated.RoomID != "room-2" {
			t.Errorf("Expected room 'room-2', got '%s'", updated.RoomID)
		}
	})

	// 3. Update Teralux (Not Found)
	// URL: PUT /api/teralux/unknown
	// SCENARIO: ID does not exist.
	// RES: 404 Not Found
	t.Run("Update Teralux (Not Found)", func(t *testing.T) {
		name := "Hack"
		req := &dtos.UpdateTeraluxRequestDTO{Name: &name}
		err := useCase.UpdateTeralux("unknown", req)
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") && !strings.Contains(err.Error(), "Teralux not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 4. Validation: Invalid Room ID
	// URL: PUT /api/teralux/t1
	// SCENARIO: Room does not exist.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Invalid Room ID", func(t *testing.T) {
		invalidRoom := "room-999"
		req := &dtos.UpdateTeraluxRequestDTO{RoomID: &invalidRoom}
		err := useCase.UpdateTeralux("t1", req)
		if err == nil {
			t.Fatal("Expected error for invalid room, got nil")
		}
		if !strings.Contains(err.Error(), "Invalid room_id: room does not exist") {
			t.Errorf("Expected 'Invalid room_id' error, got: %v", err)
		}
	})

	// 5. Validation: Empty Name (If Present)
	// URL: PUT /api/teralux/t1
	// SCENARIO: Name is empty string.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Empty Name (If Present)", func(t *testing.T) {
		emptyName := ""
		req := &dtos.UpdateTeraluxRequestDTO{Name: &emptyName}
		err := useCase.UpdateTeralux("t1", req)
		if err == nil {
			t.Fatal("Expected error for empty name, got nil")
		}
		if err.Error() != "name cannot be empty" && !strings.Contains(err.Error(), "name cannot be empty") {
			t.Errorf("Expected 'name cannot be empty', got: %v", err)
		}
	})

	// 6. Conflict: Update to Duplicate MAC
	// URL: PUT /api/teralux/t1
	// SCENARIO: MAC already taken by another device.
	// RES: 409 Conflict
	t.Run("Conflict: Update to Duplicate MAC", func(t *testing.T) {
		// Seed another device
		_ = repo.Create(&entities.Teralux{ID: "t2", MacAddress: "MAC-2", Name: "Device 2", RoomID: "r2"})

		duplicateMac := "MAC-2"
		req := &dtos.UpdateTeraluxRequestDTO{MacAddress: &duplicateMac}

		err := useCase.UpdateTeralux("t1", req)
		if err == nil {
			t.Fatal("Expected error for duplicate MAC, got nil")
		}
		if !strings.Contains(err.Error(), "Mac Address already in use") {
			t.Errorf("Expected conflict error, got: %v", err)
		}
	})

	// 7. Security: Unauthorized
}
