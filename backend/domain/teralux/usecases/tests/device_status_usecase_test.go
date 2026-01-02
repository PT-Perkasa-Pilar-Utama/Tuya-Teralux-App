package tests

import (
	"testing"

	"teralux_app/domain/teralux/dtos"

	"teralux_app/domain/teralux/entities"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"gorm.io/gorm"
)

func TestCreateDeviceStatusUseCase(t *testing.T) {
	mockRepo := &MockDeviceStatusRepository{}
	useCase := usecases.NewCreateDeviceStatusUseCase(mockRepo)

	t.Run("Success", func(t *testing.T) {
		// Mock GetByDeviceIDAndCode to return Not Found (so we can create)
		mockRepo.GetByDeviceIDAndCodeFunc = func(deviceID, code string) (*entities.DeviceStatus, error) {
			return nil, gorm.ErrRecordNotFound
		}

		mockRepo.CreateFunc = func(status *entities.DeviceStatus) error {
			if status.Code != "power_switch" {
				t.Errorf("Expected code 'power_switch', got %s", status.Code)
			}
			return nil
		}

		req := &dtos.CreateDeviceStatusRequestDTO{
			DeviceID: "device-1",
			Code:     "power_switch",
			Value:    1,
		}

		resp, err := useCase.Execute(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.ID == "" {
			t.Error("Expected ID to be set")
		}
	})

	t.Run("Duplicate Code", func(t *testing.T) {
		// Mock GetByDeviceIDAndCode to return existing record
		mockRepo.GetByDeviceIDAndCodeFunc = func(deviceID, code string) (*entities.DeviceStatus, error) {
			return &entities.DeviceStatus{ID: "existing"}, nil
		}

		req := &dtos.CreateDeviceStatusRequestDTO{
			DeviceID: "device-1",
			Code:     "power_switch",
		}

		_, err := useCase.Execute(req)

		if err == nil {
			t.Fatal("Expected error for duplicate code, got nil")
		}
		if err.Error() != "device status with this code already exists" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}
