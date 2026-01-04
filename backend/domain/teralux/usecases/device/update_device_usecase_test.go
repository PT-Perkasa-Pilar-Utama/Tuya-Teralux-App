package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestUpdateDeviceUseCase_Execute(t *testing.T) {
	repo, _ := setupDeviceTestEnv(t)
	useCase := NewUpdateDeviceUseCase(repo)

	repo.Create(&entities.Device{ID: "dev-1", Name: "Old"})

	t.Run("Success", func(t *testing.T) {
		req := &dtos.UpdateDeviceRequestDTO{Name: "New"}
		err := useCase.Execute("dev-1", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		updated, _ := repo.GetByID("dev-1")
		if updated.Name != "New" {
			t.Errorf("Expected Name 'New', got '%s'", updated.Name)
		}
	})
}
