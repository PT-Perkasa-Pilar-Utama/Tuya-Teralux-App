package usecases

import (
	"errors"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDeviceStatusByCodeUseCase_UserBehavior(t *testing.T) {
	statusRepo := new(MockDeviceStatusRepository)
	deviceRepo := new(MockDeviceRepository)
	useCase := NewGetDeviceStatusByCodeUseCase(statusRepo, deviceRepo)

	// 1. Get Device Status By Code (Success)
	t.Run("Get Device Status By Code (Success)", func(t *testing.T) {
		deviceID := "d1"
		code := "switch_1"
		deviceRepo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID}, nil).Once()
		statusRepo.On("GetByDeviceIDAndCode", deviceID, code).Return(&entities.DeviceStatus{DeviceID: deviceID, Code: code, Value: "true"}, nil).Once()

		res, err := useCase.GetDeviceStatusByCode(deviceID, code)
		assert.NoError(t, err)
		assert.Equal(t, code, res.DeviceStatus.Code)
		assert.Equal(t, "true", res.DeviceStatus.Value)
		deviceRepo.AssertExpectations(t)
		statusRepo.AssertExpectations(t)
	})

	// 2. Get Device Status By Code (Not Found)
	t.Run("Get Device Status By Code (Not Found)", func(t *testing.T) {
		deviceRepo.On("GetByID", "d1").Return(&entities.Device{ID: "d1"}, nil).Once()
		statusRepo.On("GetByDeviceIDAndCode", "d1", "unknown").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetDeviceStatusByCode("d1", "unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Status code not found")
		deviceRepo.AssertExpectations(t)
		statusRepo.AssertExpectations(t)
	})
}
