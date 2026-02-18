package usecases

import (
	"errors"
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestEnv(t *testing.T) (*repositories.TeraluxRepository, *repositories.DeviceRepository) {
	dbName := fmt.Sprintf("file:memdb_teralux_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	infrastructure.DB = db

	// Initialize config for BadgerService
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}

	err = db.AutoMigrate(&entities.Teralux{}, &entities.Device{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	tmpDir := t.TempDir()
	cache, err := infrastructure.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create Badger: %v", err)
	}

	teraluxRepo := repositories.NewTeraluxRepository(cache)
	deviceRepo := repositories.NewDeviceRepository(cache)

	return teraluxRepo, deviceRepo
}

func TestCreateTeralux_UserBehavior(t *testing.T) {
	repo, _ := setupTestEnv(t)
	useCase := NewCreateTeraluxUseCase(repo)

	// 1. Create Teralux (Success Condition)
	// URL: POST /api/teralux
	// SCENARIO: Valid payload, room exists.
	// RES: 201 Created
	t.Run("Create Teralux (Success Condition)", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:       "Master Bedroom Hub",
			MacAddress: "AA:BB:CC:11:22:33",
			RoomID:     "room-101",
		}

		res, _, err := useCase.CreateTeralux(req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.TeraluxID == "" {
			t.Error("Expected ID to be returned, got empty string")
		}
	})
	// 1b. Create Teralux with Android ID (Success Condition)
	// URL: POST /api/teralux
	// SCENARIO: Valid payload with Android ID (16 hex chars).
	// RES: 201 Created
	t.Run("Create Teralux with Android ID", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:       "Android Device",
			MacAddress: "C756630F2F039D27", // 16 chars hex
			RoomID:     "room-android",
		}

		res, _, err := useCase.CreateTeralux(req)
		if err != nil {
			t.Fatalf("Unexpected error for Android ID: %v", err)
		}
		if res.TeraluxID == "" {
			t.Error("Expected ID to be returned, got empty string")
		}
	})
	// 2. Validation: Empty Fields
	// URL: POST /api/teralux
	// SCENARIO: All fields empty.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Empty Fields", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:       "",
			MacAddress: "",
			RoomID:     "",
		}

		_, _, err := useCase.CreateTeralux(req)
		if err == nil {
			t.Fatal("Expected error for empty fields, got nil")
		}

		var valErr *utils.ValidationError
		if errors.As(err, &valErr) {
			if valErr.Message != "Validation Error" {
				t.Errorf("Expected message 'Validation Error', got '%s'", valErr.Message)
			}
			if len(valErr.Details) != 3 {
				t.Errorf("Expected 3 validation details, got %d", len(valErr.Details))
			}
		} else {
			t.Fatalf("Expected utils.ValidationError, got %T: %v", err, err)
		}
	})

	// 3. Validation: Invalid MAC Address Format
	// URL: POST /api/teralux
	// SCENARIO: Invalid MAC format.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Invalid MAC Address Format", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:       "Living Room",
			MacAddress: "INVALID-MAC",
			RoomID:     "room-1",
		}

		_, _, err := useCase.CreateTeralux(req)
		if err == nil {
			t.Fatal("Expected error for invalid mac, got nil")
		}

		var valErr *utils.ValidationError
		if errors.As(err, &valErr) {
			found := false
			for _, d := range valErr.Details {
				if d.Field == "mac_address" && d.Message == "invalid mac address format" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected detail for mac_address invalid format")
			}
		}
	})

	// 4. Validation: Name Too Long
	// URL: POST /api/teralux
	// SCENARIO: Name > 255 chars.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Name Too Long", func(t *testing.T) {
		longName := strings.Repeat("a", 256)
		req := &dtos.CreateTeraluxRequestDTO{
			Name:       longName,
			MacAddress: "AA:BB:CC:11:22:33",
			RoomID:     "room-101",
		}

		_, _, err := useCase.CreateTeralux(req)
		if err == nil {
			t.Fatal("Expected error for long name, got nil")
		}

		var valErr *utils.ValidationError
		if errors.As(err, &valErr) {
			found := false
			for _, d := range valErr.Details {
				if d.Field == "name" && d.Message == "name must be 255 characters or less" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected detail for name too long")
			}
		}
	})

	// 5. Idempotent: Duplicate MAC Address Returns Existing ID
	// URL: POST /api/teralux
	// SCENARIO: MAC already exists (idempotent for booting).
	// RES: 200 OK
	t.Run("Idempotent: Duplicate MAC Address Returns Existing ID", func(t *testing.T) {
		_ = repo.Create(&entities.Teralux{ID: "existing-id", MacAddress: "DD:EE:FF:11:22:33", RoomID: "r1", Name: "Existing"})

		req := &dtos.CreateTeraluxRequestDTO{
			Name:       "New Hub",
			MacAddress: "DD:EE:FF:11:22:33",
			RoomID:     "room-102",
		}

		res, _, err := useCase.CreateTeralux(req)
		if err != nil {
			t.Fatalf("Expected no error for duplicate MAC (idempotent), got: %v", err)
		}
		if res.TeraluxID != "existing-id" {
			t.Errorf("Expected existing ID 'existing-id', got: %s", res.TeraluxID)
		}
	})

	// 6. Security: Unauthorized (Missing Auth)
}
