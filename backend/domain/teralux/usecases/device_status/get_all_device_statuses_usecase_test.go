package usecases

import (
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper for device_status package tests
func setupStatusTestEnv(t *testing.T) (*repositories.DeviceStatusRepository, *repositories.DeviceRepository) {
	dbName := fmt.Sprintf("file:memdb_status_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	infrastructure.DB = db

	// Initialize config for BadgerService
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}

	err = db.AutoMigrate(&entities.DeviceStatus{}, &entities.Device{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	tmpDir := t.TempDir()
	cache, err := infrastructure.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create Badger: %v", err)
	}
	return repositories.NewDeviceStatusRepository(cache), repositories.NewDeviceRepository(cache)
}

func TestGetAllDeviceStatusesUseCase_UserBehavior(t *testing.T) {
	repo, _ := setupStatusTestEnv(t)
	useCase := NewGetAllDeviceStatusesUseCase(repo)

	// Seed
	_ = repo.Upsert(&entities.DeviceStatus{DeviceID: "d1", Code: "c1", Value: "v1"})
	_ = repo.Upsert(&entities.DeviceStatus{DeviceID: "d2", Code: "c2", Value: "v2"})

	// 1. Get All Statuses (Success)
	// URL: GET /api/device-statuses
	// RES: 200 OK
	t.Run("Get All Statuses (Success)", func(t *testing.T) {
		res, err := useCase.Execute(0, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res.DeviceStatuses) < 2 {
			t.Errorf("Expected at least 2 statuses, got %d", len(res.DeviceStatuses))
		}
	})

	// 2. Get All Statuses (Empty)
	// URL: GET /api/device-statuses
	// RES: 200 OK
	t.Run("Get All Statuses (Empty)", func(t *testing.T) {
		// Create new empty env
		emptyRepo, _ := setupStatusTestEnv(t)
		emptyUC := NewGetAllDeviceStatusesUseCase(emptyRepo)

		res, err := emptyUC.Execute(0, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res.DeviceStatuses) != 0 {
			t.Errorf("Expected 0 statuses, got %d", len(res.DeviceStatuses))
		}
	})

	// 3. Unauthorized (Middleware)
}
