package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDeviceStatusesByDeviceIDUseCase_UserBehavior(t *testing.T) {
	statusRepo := new(MockDeviceStatusRepository)
	deviceRepo := new(MockDeviceRepository)
	useCase := NewGetDeviceStatusesByDeviceIDUseCase(statusRepo, deviceRepo)

	// 1. Get Device Statuses By Device ID (Success)
	t.Run("Get Device Statuses By Device ID (Success)", func(t *testing.T) {
		deviceID := "d1"
		deviceRepo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID}, nil).Once()
		statusRepo.On("GetByDeviceIDPaginated", deviceID, 0, 10).Return([]entities.DeviceStatus{
			{DeviceID: deviceID, Code: "c1", Value: "v1"},
			{DeviceID: deviceID, Code: "c2", Value: "v2"},
		}, int64(2), nil).Once()

		res, err := useCase.ListDeviceStatusesByDeviceID(deviceID, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 2, res.Total)
		assert.Len(t, res.DeviceStatuses, 2)
		deviceRepo.AssertExpectations(t)
		statusRepo.AssertExpectations(t)
	})

	// 2. Get Device Statuses By Device ID (Invalid Device)
	t.Run("Get Device Statuses By Device ID (Invalid Device)", func(t *testing.T) {
		deviceRepo.On("GetByID", "unknown").Return(nil, assert.AnError).Once()

		_, err := useCase.ListDeviceStatusesByDeviceID("unknown", 1, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Device not found")
		deviceRepo.AssertExpectations(t)
	})
}
