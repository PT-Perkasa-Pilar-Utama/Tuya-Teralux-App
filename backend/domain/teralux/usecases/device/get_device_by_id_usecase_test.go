package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetDeviceByIDUseCase_Execute(t *testing.T) {
	repo, _ := setupDeviceTestEnv(t)
	useCase := NewGetDeviceByIDUseCase(repo)

	repo.Create(&entities.Device{ID: "dev-123", Name: "Test Device"})

	t.Run("Success - Device found", func(t *testing.T) {
		res, err := useCase.Execute("dev-123")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if res.ID != "dev-123" {
			t.Errorf("Expected ID 'dev-123', got '%s'", res.ID)
		}
	})

	t.Run("Error - Device not found", func(t *testing.T) {
		_, err := useCase.Execute("unknown")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}
