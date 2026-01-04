package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"
)

func TestGetAllDevicesUseCase_Execute(t *testing.T) {
	repo, _ := setupDeviceTestEnv(t)
	useCase := NewGetAllDevicesUseCase(repo)

	repo.Create(&entities.Device{ID: "dev-1", Name: "D1"})
	repo.Create(&entities.Device{ID: "dev-2", Name: "D2"})

	t.Run("Success - Returns all devices", func(t *testing.T) {
		res, err := useCase.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(res.Devices) < 2 {
			t.Errorf("Expected at least 2 devices, got %d", len(res.Devices))
		}
	})
}
