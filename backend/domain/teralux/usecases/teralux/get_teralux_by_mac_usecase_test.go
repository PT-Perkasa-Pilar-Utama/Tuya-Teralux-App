package usecases

import (
	"strings"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetTeraluxByMAC_UserBehavior(t *testing.T) {
	repo, _ := setupTestEnv(t)
	useCase := NewGetTeraluxByMACUseCase(repo)

	// Seed data
	repo.Create(&entities.Teralux{ID: "t1", Name: "Living Room", MacAddress: "AA:BB:CC:11:22:33"})

	// 1. Get Teralux By MAC (Success)
	// URL: GET /api/teralux/mac/AA:BB:CC:11:22:33
	// SCENARIO: Device valid mac.
	// RES: 200 OK
	t.Run("Get Teralux By MAC (Success)", func(t *testing.T) {
		res, err := useCase.Execute("AA:BB:CC:11:22:33")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Teralux.MacAddress != "AA:BB:CC:11:22:33" {
			t.Errorf("Expected MAC 'AA:BB:CC:11:22:33', got '%s'", res.Teralux.MacAddress)
		}
	})

	// 2. Get Teralux By MAC (Not Found)
	// URL: GET /api/teralux/mac/XX:YY:ZZ:00:00:00
	// SCENARIO: MAC does not exist.
	// RES: 404 Not Found
	t.Run("Get Teralux By MAC (Not Found)", func(t *testing.T) {
		_, err := useCase.Execute("XX:YY:ZZ:00:00:00")
		if err == nil {
			t.Fatal("Expected error for unknown MAC, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") && !strings.Contains(err.Error(), "Teralux not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 3. Validation: Invalid MAC Format
	// URL: GET /api/teralux/mac/INVALID-MAC
	// SCENARIO: MAC string is invalid.
	// RES: 400 Bad Request
	t.Run("Validation: Invalid MAC Format", func(t *testing.T) {
		_, err := useCase.Execute("INVALID-MAC")
		if err == nil {
			t.Fatal("Expected error for invalid mac, got nil")
		}
		if err.Error() != "Invalid MAC address format" {
			t.Errorf("Expected 'Invalid MAC address format', got: %v", err)
		}
	})

	// 4. Security: Unauthorized
}
