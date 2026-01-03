package tests

import (
	"testing"

	"teralux_app/domain/teralux/dtos"

	"teralux_app/domain/teralux/entities"
	usecases "teralux_app/domain/teralux/usecases/device_status"
)

func TestGetDeviceStatusesByDeviceIDUseCase(t *testing.T) {
	mockRepo := &MockDeviceStatusRepository{}
	useCase := usecases.NewGetDeviceStatusesByDeviceIDUseCase(mockRepo)

	t.Run("Success", func(t *testing.T) {
		deviceID := "device-123"
		mockRepo.GetByDeviceIDFunc = func(id string) ([]entities.DeviceStatus, error) {
			return []entities.DeviceStatus{
				{DeviceID: deviceID, Code: "switch_1", Value: "true"},
				{DeviceID: deviceID, Code: "switch_2", Value: "false"},
			}, nil
		}

		resp, err := useCase.Execute(deviceID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("Expected 2 statuses, got %d", len(resp))
		}
	})
}

func TestGetAllDeviceStatusesUseCase(t *testing.T) {
	mockRepo := &MockDeviceStatusRepository{}
	useCase := usecases.NewGetAllDeviceStatusesUseCase(mockRepo)

	t.Run("Success", func(t *testing.T) {
		mockRepo.GetAllFunc = func() ([]entities.DeviceStatus, error) {
			return []entities.DeviceStatus{
				{DeviceID: "d-1", Code: "c-1", Value: "v-1"},
				{DeviceID: "d-2", Code: "c-2", Value: "v-2"},
			}, nil
		}

		resp, err := useCase.Execute()

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(resp.Statuses) != 2 {
			t.Errorf("Expected 2 statuses, got %d", len(resp.Statuses))
		}
	})
}

func TestGetDeviceStatusByCodeUseCase(t *testing.T) {
	mockRepo := &MockDeviceStatusRepository{}
	useCase := usecases.NewGetDeviceStatusByCodeUseCase(mockRepo)

	t.Run("Found", func(t *testing.T) {
		deviceID := "device-1"
		code := "switch_1"
		mockRepo.GetByDeviceIDAndCodeFunc = func(d, c string) (*entities.DeviceStatus, error) {
			return &entities.DeviceStatus{DeviceID: d, Code: c, Value: "true"}, nil
		}

		resp, err := useCase.Execute(deviceID, code)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.Code != code {
			t.Errorf("Expected code %s, got %s", code, resp.Code)
		}
	})
}

func TestUpdateDeviceStatusUseCase(t *testing.T) {
	mockRepo := &MockDeviceStatusRepository{}
	useCase := usecases.NewUpdateDeviceStatusUseCase(mockRepo)

	t.Run("Success", func(t *testing.T) {
		deviceID := "device-1"
		code := "switch_1"
		newValue := "false"

		mockRepo.GetByDeviceIDAndCodeFunc = func(d, c string) (*entities.DeviceStatus, error) {
			return &entities.DeviceStatus{DeviceID: d, Code: c, Value: "true"}, nil
		}
		mockRepo.UpsertFunc = func(status *entities.DeviceStatus) error {
			if status.Value != newValue {
				t.Errorf("Expected value %s, got %s", newValue, status.Value)
			}
			return nil
		}

		req := &dtos.UpdateDeviceStatusRequestDTO{Value: newValue}
		err := useCase.Execute(deviceID, code, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}
