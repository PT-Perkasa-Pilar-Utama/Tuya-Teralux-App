package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetAllTeralux_UserBehavior(t *testing.T) {
	repo, _ := setupTestEnv(t)
	useCase := NewGetAllTeraluxUseCase(repo)

	// 1. Get All Teralux (Success - Empty List)
	// URL: GET /api/teralux
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "message": "...", "data": { "teralux": [], "total": 0 } }
	t.Run("Get All Teralux (Success - Empty List)", func(t *testing.T) {
		res, err := useCase.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Total != 0 {
			t.Errorf("Expected total 0, got %d", res.Total)
		}
		if len(res.Teralux) != 0 {
			t.Errorf("Expected empty list, got %d items", len(res.Teralux))
		}
	})

	// Seed data
	repo.Create(&entities.Teralux{ID: "t1", Name: "T1", MacAddress: "M1"})
	repo.Create(&entities.Teralux{ID: "t2", Name: "T2", MacAddress: "M2"})

	// 2. Get All Teralux (Success - With Data)
	// URL: GET /api/teralux
	// METHOD: GET
	// RES: 200 OK
	// RESPONSE: { "status": true, "message": "...", "data": { "teralux": [ ... ], "total": 2 } }
	t.Run("Get All Teralux (Success - With Data)", func(t *testing.T) {
		res, err := useCase.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Total != 2 {
			t.Errorf("Expected total 2, got %d", res.Total)
		}
		if len(res.Teralux) != 2 {
			t.Errorf("Expected 2 items, got %d", len(res.Teralux))
		}
	})
}
