package usecases

import (
	"fmt"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetAllTeralux_UserBehavior(t *testing.T) {
	repo, _, _ := setupTestEnv(t)
	useCase := NewGetAllTeraluxUseCase(repo)

	// 1. Get All Teralux (Success - Empty List)
	// URL: GET /api/teralux
	// SCENARIO: No data.
	// RES: 200 OK
	t.Run("Get All Teralux (Success - Empty List)", func(t *testing.T) {
		filter := &dtos.TeraluxFilterDTO{Page: 1, Limit: 10}
		res, err := useCase.ListTeralux(filter)
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
	_ = repo.Create(&entities.Teralux{ID: "t1", Name: "Hub 1", MacAddress: "M1", RoomID: "r1"})
	_ = repo.Create(&entities.Teralux{ID: "t2", Name: "Hub 2", MacAddress: "M2", RoomID: "r2"})
	// Add more for pagination
	for i := 3; i <= 15; i++ {
		_ = repo.Create(&entities.Teralux{ID: fmt.Sprintf("t%d", i), Name: fmt.Sprintf("Hub %d", i), MacAddress: fmt.Sprintf("M%d", i), RoomID: "r3"})
	}

	// 2. Get All Teralux (Success - With Data)
	// URL: GET /api/teralux
	// SCENARIO: Has data, default pagination.
	// RES: 200 OK
	t.Run("Get All Teralux (Success - With Data)", func(t *testing.T) {
		filter := &dtos.TeraluxFilterDTO{Page: 1, Limit: 10}
		res, err := useCase.ListTeralux(filter)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// Total should be 15
		if res.Total != 15 {
			t.Errorf("Expected total 15, got %d", res.Total)
		}
		// Default limit 10
		if len(res.Teralux) != 10 {
			t.Errorf("Expected 10 items, got %d", len(res.Teralux))
		}
	})

	// 3. Pagination: Limit and Page
	// URL: GET /api/teralux?page=2&limit=5
	// SCENARIO: Specific page/limit.
	// RES: 200 OK
	t.Run("Pagination: Limit and Page", func(t *testing.T) {
		filter := &dtos.TeraluxFilterDTO{Page: 2, Limit: 5}
		res, err := useCase.ListTeralux(filter)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Total != 15 {
			t.Errorf("Expected total 15, got %d", res.Total)
		}
		if len(res.Teralux) != 5 {
			t.Errorf("Expected 5 items, got %d", len(res.Teralux))
		}
		if res.Page != 2 {
			t.Errorf("Expected page 2, got %d", res.Page)
		}
		if res.PerPage != 5 {
			t.Errorf("Expected per_page 5, got %d", res.PerPage)
		}
	})

	// 4. Filter: By Room ID
	// URL: GET /api/teralux?room_id=room-101
	// SCENARIO: Filter by room.
	// RES: 200 OK
	t.Run("Filter: By Room ID", func(t *testing.T) {
		// Only t1 has room r1.
		roomID := "r1"
		filter := &dtos.TeraluxFilterDTO{Page: 1, Limit: 10, RoomID: &roomID}
		res, err := useCase.ListTeralux(filter)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.Total != 1 {
			t.Errorf("Expected total 1, got %d", res.Total)
		}
		if len(res.Teralux) != 1 {
			t.Errorf("Expected 1 item, got %d", len(res.Teralux))
		}
		if res.Teralux[0].RoomID != "r1" {
			t.Errorf("Expected room 'r1', got '%s'", res.Teralux[0].RoomID)
		}
	})

	// 5. Security: Unauthorized
}
