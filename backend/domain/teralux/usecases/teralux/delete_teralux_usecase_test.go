package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestDeleteTeralux_UserBehavior(t *testing.T) {
	repo, _ := setupTestEnv(t)
	useCase := NewDeleteTeraluxUseCase(repo)

	// Seed data
	t1 := &entities.Teralux{ID: "tx-1", Name: "T1", MacAddress: "AA:BB"}
	_ = repo.Create(t1)

	// 1. Delete Teralux (Success Condition)
	// URL: DELETE /api/teralux/tx-1
	// SCENARIO: Device exists.
	// RES: 200 OK
	t.Run("Delete Teralux (Success Condition)", func(t *testing.T) {
		err := useCase.Execute("tx-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify deletion
		_, err = repo.GetByID("tx-1")
		if err == nil {
			t.Error("Expected error getting deleted item, got nil")
		}
	})

	// 2. Delete Teralux (Not Found)
	// URL: DELETE /api/teralux/tx-999
	// SCENARIO: Device does not exist.
	// RES: 404 Not Found
	t.Run("Delete Teralux (Not Found)", func(t *testing.T) {
		err := useCase.Execute("tx-999")
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "Teralux not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 3. Validation: Invalid ID Format
	// URL: DELETE /api/teralux/INVALID-UUID
	// SCENARIO: ID is not a valid UUID format.
	// RES: 400 Bad Request
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		err := useCase.Execute("INVALID-UUID")
		if err == nil {
			t.Fatal("Expected error for invalid ID format, got nil")
		}
		if err.Error() != "Invalid ID format" {
			t.Errorf("Expected 'Invalid ID format', got: %v", err)
		}
	})

	// 4. Security: Unauthorized
}
