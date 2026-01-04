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
	repo.Create(&entities.Teralux{ID: "t1", Name: "Living Room", MacAddress: "AA:BB:CC"})

	// 1. Get Teralux By MAC (Success)
	// URL: GET /api/teralux/mac/AA:BB:CC
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "data": { "id": "t1", "mac_address": "AA:BB:CC" } }
	t.Run("Get Teralux By MAC (Success)", func(t *testing.T) {
		res, err := useCase.Execute("AA:BB:CC")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.MacAddress != "AA:BB:CC" {
			t.Errorf("Expected MAC 'AA:BB:CC', got '%s'", res.MacAddress)
		}
	})

	// 2. Get Teralux By MAC (Not Found)
	// URL: GET /api/teralux/mac/XX:YY:ZZ
	// METHOD: GET
	// RES: 404 Not Found
	// RESPONSE: { "status": false, "message": "record not found", "data": nil }
	t.Run("Get Teralux By MAC (Not Found)", func(t *testing.T) {
		_, err := useCase.Execute("XX:YY:ZZ")
		if err == nil {
			t.Fatal("Expected error for unknown MAC, got nil")
		}
		if !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'record not found', got: %v", err)
		}
	})
}
