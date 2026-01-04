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

// Helper for device package tests
func setupDeviceTestEnv(t *testing.T) (*repositories.DeviceRepository, *repositories.DeviceStatusRepository) {
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

	err = db.AutoMigrate(&entities.Device{}, &entities.DeviceStatus{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	tmpDir := t.TempDir()
	cache, err := persistence.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create Badger: %v", err)
	}
	return repositories.NewDeviceRepository(cache), repositories.NewDeviceStatusRepository(cache)
}

func TestDeleteDeviceUseCase_Execute(t *testing.T) {
	repo, statusRepo := setupDeviceTestEnv(t)
	useCase := NewDeleteDeviceUseCase(repo, statusRepo)

	repo.Create(&entities.Device{ID: "dev-1", Name: "To Delete"})

	t.Run("Success", func(t *testing.T) {
		err := useCase.Execute("dev-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		_, err = repo.GetByID("dev-1")
		if err == nil {
			t.Error("Expected error for deleted device, got nil")
		}
	})
}
