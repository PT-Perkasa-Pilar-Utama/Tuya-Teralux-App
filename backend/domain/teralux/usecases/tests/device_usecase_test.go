package tests

import (
	"errors"
	"testing"
	"time"

	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	usecases "teralux_app/domain/teralux/usecases/device"
	tuya_dtos "teralux_app/domain/tuya/dtos"
)

func TestCreateDeviceUseCase(t *testing.T) {
	mockRepo := &MockDeviceRepository{}
	mockStatusRepo := &MockDeviceStatusRepository{}
	mockTuyaAuth := &MockTuyaAuthUseCase{}
	mockTuyaGetDevice := &MockTuyaGetDeviceByIDUseCase{}

	useCase := usecases.NewCreateDeviceUseCase(mockRepo, mockStatusRepo, mockTuyaAuth, mockTuyaGetDevice)

	t.Run("Success with Tuya Data", func(t *testing.T) {
		tuyaID := "tuya-123"
		accessToken := "token-abc"

		mockTuyaAuth.AuthenticateFunc = func() (*tuya_dtos.TuyaAuthResponseDTO, error) {
			return &tuya_dtos.TuyaAuthResponseDTO{AccessToken: accessToken}, nil
		}

		mockTuyaGetDevice.GetDeviceByIDFunc = func(token, id string) (*tuya_dtos.TuyaDeviceDTO, error) {
			if id != tuyaID {
				t.Errorf("Expected tuyaID %s, got %s", tuyaID, id)
			}
			return &tuya_dtos.TuyaDeviceDTO{
				ID:       id,
				Name:     "Tuya Name",
				Category: "switch",
				Status: []tuya_dtos.TuyaDeviceStatusDTO{
					{Code: "switch_1", Value: true},
				},
			}, nil
		}

		mockRepo.GetByTeraluxIDFunc = func(teraluxID string) ([]entities.Device, error) {
			return []entities.Device{}, nil
		}
		mockRepo.CreateFunc = func(device *entities.Device) error {
			if device.Category != "switch" {
				t.Errorf("Expected category 'switch', got %s", device.Category)
			}
			return nil
		}

		statusUpserted := false
		mockStatusRepo.UpsertDeviceStatusesFunc = func(deviceID string, statuses []entities.DeviceStatus) error {
			if len(statuses) != 1 {
				t.Errorf("Expected 1 status, got %d", len(statuses))
			}
			if statuses[0].Value != "true" {
				t.Errorf("Expected value 'true', got %s", statuses[0].Value)
			}
			statusUpserted = true
			return nil
		}

		req := &dtos.CreateDeviceRequestDTO{
			ID:        tuyaID,
			TeraluxID: "teralux-123",
			Name:      "Local Name",
		}

		resp, err := useCase.Execute(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.ID == "" {
			t.Error("Expected generated ID, got empty")
		}
		if !statusUpserted {
			t.Error("Expected status to be upserted from Tuya")
		}
	})

	t.Run("Tuya Auth Failure", func(t *testing.T) {
		mockTuyaAuth.AuthenticateFunc = func() (*tuya_dtos.TuyaAuthResponseDTO, error) {
			return nil, errors.New("auth failed")
		}

		req := &dtos.CreateDeviceRequestDTO{
			ID:        "tuya-123",
			TeraluxID: "teralux-123",
			Name:      "Local Name",
		}

		_, err := useCase.Execute(req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

func TestDeleteDeviceUseCase(t *testing.T) {
	mockRepo := &MockDeviceRepository{}
	mockStatusRepo := &MockDeviceStatusRepository{}
	useCase := usecases.NewDeleteDeviceUseCase(mockRepo, mockStatusRepo)

	t.Run("Success Cleanup", func(t *testing.T) {
		deviceID := "device-123"

		statusDeleted := false
		mockStatusRepo.DeleteByDeviceIDFunc = func(id string) error {
			if id != deviceID {
				t.Errorf("Expected id %s, got %s", deviceID, id)
			}
			statusDeleted = true
			return nil
		}

		deviceDeleted := false
		mockRepo.DeleteFunc = func(id string) error {
			if id != deviceID {
				t.Errorf("Expected id %s, got %s", deviceID, id)
			}
			deviceDeleted = true
			return nil
		}

		err := useCase.Execute(deviceID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !statusDeleted {
			t.Error("Expected statuses to be deleted")
		}
		if !deviceDeleted {
			t.Error("Expected device to be deleted")
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
