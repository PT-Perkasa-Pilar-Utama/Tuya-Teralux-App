package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper for device package tests
func setupDeviceTestEnv(t *testing.T) (*repositories.DeviceRepository, *repositories.DeviceStatusRepository, *repositories.TeraluxRepository) {
	dbName := fmt.Sprintf("file:memdb_dev_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	infrastructure.DB = db

	// Initialize config for BadgerService
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}

	err = db.AutoMigrate(&entities.Device{}, &entities.DeviceStatus{}, &entities.Teralux{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	tmpDir := t.TempDir()
	cache, err := infrastructure.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create Badger: %v", err)
	}
	return repositories.NewDeviceRepository(cache), repositories.NewDeviceStatusRepository(cache), repositories.NewTeraluxRepository(cache)
}

func TestDeleteDeviceUseCase_UserBehavior(t *testing.T) {
	repo, statusRepo, teraRepo := setupDeviceTestEnv(t)
	useCase := NewDeleteDeviceUseCase(repo, statusRepo, teraRepo)

	_ = repo.Create(&entities.Device{ID: "dev-1", Name: "To Delete", TeraluxID: "tx-1"})

	// 1. Delete Device (Success)
	// URL: DELETE /api/devices/dev-1
	// SCENARIO: Device exists.
	// RES: 200 OK
	t.Run("Delete Device (Success)", func(t *testing.T) {
		err := useCase.Execute("dev-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify deletion
		_, err = repo.GetByID("dev-1")
		if err == nil {
			t.Error("Expected error for deleted device, got nil")
		}
	})

	// 2. Delete Device (Not Found)
	// URL: DELETE /api/devices/dev-999
	// SCENARIO: Device does not exist.
	// RES: 404 Not Found
	t.Run("Delete Device (Not Found)", func(t *testing.T) {
		err := useCase.Execute("dev-999")
		if err == nil {
			t.Fatal("Expected error for unknown ID, got nil")
		}
		if !strings.Contains(err.Error(), "Device not found") && !strings.Contains(err.Error(), "record not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	// 3. Validation: Invalid ID Format
	// URL: DELETE /api/devices/INVALID
	// SCENARIO: Invalid UUID/ID format.
	// RES: 400 Bad Request
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		err := useCase.Execute("INVALID")
		if err == nil {
			t.Fatal("Expected error for invalid ID, got nil")
		}
		if err.Error() != "Invalid ID format" {
			t.Errorf("Expected 'Invalid ID format', got: %v", err)
		}
	})

	// 4. Security: Unauthorized
}
