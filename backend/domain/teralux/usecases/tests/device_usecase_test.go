package tests

import (
	"errors"
	"testing"
	"time"

	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	usecases "teralux_app/domain/teralux/usecases/device"
)

func TestCreateDeviceUseCase(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockDeviceRepository{}
		useCase := usecases.NewCreateDeviceUseCase(mockRepo)

		mockRepo.GetByTeraluxIDFunc = func(teraluxID string) ([]entities.Device, error) {
			return []entities.Device{}, nil
		}
		mockRepo.CreateFunc = func(device *entities.Device) error {
			if device.Name != "Test Device" {
				t.Errorf("Expected name 'Test Device', got %s", device.Name)
			}
			device.ID = "generated-id"
			return nil
		}

		req := &dtos.CreateDeviceRequestDTO{
			TeraluxID: "teralux-123",
			Name:      "Test Device",
		}

		resp, err := useCase.Execute(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.ID == "" {
			t.Error("Expected generated ID, got empty")
		}
	})

	t.Run("Device Already Exists", func(t *testing.T) {
		mockRepo := &MockDeviceRepository{}
		useCase := usecases.NewCreateDeviceUseCase(mockRepo)

		expectedID := "existing-device-id"
		mockRepo.GetByTeraluxIDFunc = func(teraluxID string) ([]entities.Device, error) {
			return []entities.Device{{
				ID:        expectedID,
				TeraluxID: teraluxID,
				Name:      "Existing Device",
			}}, nil
		}

		req := &dtos.CreateDeviceRequestDTO{
			TeraluxID: "teralux-123",
			Name:      "New Name (Ignored)",
		}

		resp, err := useCase.Execute(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.ID != expectedID {
			t.Errorf("Expected ID %s, got %s", expectedID, resp.ID)
		}
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := &MockDeviceRepository{}
		useCase := usecases.NewCreateDeviceUseCase(mockRepo)

		mockRepo.CreateFunc = func(device *entities.Device) error {
			return errors.New("db error")
		}

		req := &dtos.CreateDeviceRequestDTO{
			TeraluxID: "teralux-123",
			Name:      "Test Device",
		}

		_, err := useCase.Execute(req)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != "db error" {
			t.Errorf("Expected 'db error', got %v", err)
		}
	})
}

func TestGetDeviceByIDUseCase(t *testing.T) {
	mockRepo := &MockDeviceRepository{}
	useCase := usecases.NewGetDeviceByIDUseCase(mockRepo)

	t.Run("Found", func(t *testing.T) {
		expectedID := "device-123"
		mockRepo.GetByIDFunc = func(id string) (*entities.Device, error) {
			if id != expectedID {
				t.Errorf("Expected ID %s, got %s", expectedID, id)
			}
			return &entities.Device{
				ID:        expectedID,
				Name:      "Found Device",
				CreatedAt: time.Now(),
			}, nil
		}

		resp, err := useCase.Execute(expectedID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.ID != expectedID {
			t.Errorf("Expected ID %s, got %s", expectedID, resp.ID)
		}
		if resp.Name != "Found Device" {
			t.Errorf("Expected Name 'Found Device', got %s", resp.Name)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.GetByIDFunc = func(id string) (*entities.Device, error) {
			return nil, errors.New("record not found")
		}

		_, err := useCase.Execute("unknown")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}
