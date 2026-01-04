package usecases

import (
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper for device_status package tests
func setupStatusTestEnv(t *testing.T) *repositories.DeviceStatusRepository {
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

	err = db.AutoMigrate(&entities.DeviceStatus{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	tmpDir := t.TempDir()
	cache, err := persistence.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create Badger: %v", err)
	}
	return repositories.NewDeviceStatusRepository(cache)
}

func TestGetAllDeviceStatusesUseCase_Execute(t *testing.T) {
	repo := setupStatusTestEnv(t)
	useCase := NewGetAllDeviceStatusesUseCase(repo)

	// Seed
	repo.Upsert(&entities.DeviceStatus{DeviceID: "d1", Code: "c1", Value: "v1"})
	repo.Upsert(&entities.DeviceStatus{DeviceID: "d2", Code: "c2", Value: "v2"})

	t.Run("Success", func(t *testing.T) {
		res, err := useCase.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res.Statuses) < 2 {
			t.Errorf("Expected at least 2 statuses, got %d", len(res.Statuses))
		}
	})
}
