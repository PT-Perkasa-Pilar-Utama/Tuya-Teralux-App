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
	"teralux_app/domain/teralux/services"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestEnv(t *testing.T) (*repositories.TeraluxRepository, *repositories.DeviceRepository, *services.TeraluxExternalService) {
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
	externalService := services.NewTeraluxExternalService()

	return teraluxRepo, deviceRepo, externalService
}

func TestCreateTeralux_UserBehavior(t *testing.T) {
	repo, _, extSvc := setupTestEnv(t)
	useCase := NewCreateTeraluxUseCase(repo, extSvc)

	// 1. Create Teralux (Success Condition)
	// URL: POST /api/teralux
	// SCENARIO: Valid payload, room exists.
	// RES: 201 Created
	t.Run("Create Teralux (Success Condition)", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:       "Master Bedroom Hub",
			MacAddress:   "AA:BB:CC:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
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
			MacAddress:   "C756630F2F039D27", // 16 chars hex
			RoomID:       "1",
			DeviceTypeID: "1",
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
			MacAddress:   "",
			RoomID:       "",
			DeviceTypeID: "",
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
			if len(valErr.Details) != 4 {
				t.Errorf("Expected 4 validation details, got %d", len(valErr.Details))
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
			MacAddress:   "INVALID-MAC",
			RoomID:       "1",
			DeviceTypeID: "1",
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
			MacAddress:   "AA:BB:CC:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
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

	// 5. Conflict: Duplicate MAC Address Returns 409 (With Normalization)
	// URL: POST /api/teralux
	// SCENARIO: MAC already exists. Input is lowercase, but repo has uppercase.
	// RES: 409 Conflict
	t.Run("Conflict: Duplicate MAC Address Returns 409 (With Normalization)", func(t *testing.T) {
		_ = repo.Create(&entities.Teralux{ID: "existing-id", MacAddress: "DD:EE:FF:11:22:33", RoomID: "r1", Name: "Existing"})

		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "New Hub",
			MacAddress:   "dd:ee:ff:11:22:33", // Lowercase
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		_, _, err := useCase.CreateTeralux(req)
		if err == nil {
			t.Fatal("Expected error for duplicate MAC, got nil")
		}

		var apiErr *utils.APIError
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode != 409 {
				t.Errorf("Expected status 409, got %d", apiErr.StatusCode)
			}
		} else {
			t.Fatalf("Expected utils.APIError, got %T: %v", err, err)
		}
	})

	// 5b. Create Teralux with 12-digit RAW MAC (Success)
	// URL: POST /api/teralux
	// SCENARIO: Valid 12-digit raw MAC.
	// RES: 201 Created
	t.Run("Create Teralux with 12-digit RAW MAC", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "Raw Hub",
			MacAddress:   "aabbcc112233",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		res, _, err := useCase.CreateTeralux(req)
		if err != nil {
			t.Fatalf("Unexpected error for 12-digit MAC: %v", err)
		}

		// Verify it was normalized to uppercase in DB
		saved, _ := repo.GetByID(res.TeraluxID)
		if saved.MacAddress != "AABBCC112233" {
			t.Errorf("Expected normalized MAC 'AABBCC112233', got '%s'", saved.MacAddress)
		}
	})

	// 6. Security: Unauthorized (Missing Auth)
}
