package tests

import (
	"errors"
	"testing"
	"time"

	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	usecases "teralux_app/domain/teralux/usecases/teralux"
)

func TestCreateTeraluxUseCase(t *testing.T) {
	mockRepo := &MockTeraluxRepository{}
	useCase := usecases.NewCreateTeraluxUseCase(mockRepo)

	t.Run("Success", func(t *testing.T) {
		mockRepo.CreateFunc = func(teralux *entities.Teralux) error {
			if teralux.Name != "Test Teralux" {
				t.Errorf("Expected name 'Test Teralux', got %s", teralux.Name)
			}
			teralux.ID = "generated-id"
			return nil
		}

		req := &dtos.CreateTeraluxRequestDTO{
			MacAddress: "AA:BB:CC:DD:EE:FF",
			Name:       "Test Teralux",
		}

		resp, err := useCase.Execute(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.ID == "" {
			t.Error("Expected generated ID, got empty")
		}
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.CreateFunc = func(teralux *entities.Teralux) error {
			return errors.New("db error")
		}

		req := &dtos.CreateTeraluxRequestDTO{
			MacAddress: "AA:BB:CC:DD:EE:FF",
			Name:       "Test Teralux",
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

func TestGetTeraluxByIDUseCase(t *testing.T) {
	mockRepo := &MockTeraluxRepository{}
	mockDeviceRepo := &MockDeviceRepository{}
	useCase := usecases.NewGetTeraluxByIDUseCase(mockRepo, mockDeviceRepo)

	t.Run("Found", func(t *testing.T) {
		expectedID := "teralux-123"
		mockRepo.GetByIDFunc = func(id string) (*entities.Teralux, error) {
			if id != expectedID {
				t.Errorf("Expected ID %s, got %s", expectedID, id)
			}
			return &entities.Teralux{
				ID:        expectedID,
				Name:      "Found Teralux",
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
		if resp.Name != "Found Teralux" {
			t.Errorf("Expected Name 'Found Teralux', got %s", resp.Name)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.GetByIDFunc = func(id string) (*entities.Teralux, error) {
			return nil, errors.New("record not found")
		}

		_, err := useCase.Execute("unknown")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}
