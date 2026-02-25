package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllDevicesUseCase_UserBehavior(t *testing.T) {
	// 1. Get All Devices (Success - Filter by Teralux)
	t.Run("Get All Devices (Success - Filter by Teralux)", func(t *testing.T) {
		repo := new(MockDeviceRepository)
		useCase := NewGetAllDevicesUseCase(repo)
		teraID := "tx-1"
		filter := &dtos.DeviceFilterDTO{TeraluxID: &teraID}
		expectedDevices := []entities.Device{
			{ID: "d1", Name: "Light 1", TeraluxID: "tx-1"},
			{ID: "d2", Name: "Light 2", TeraluxID: "tx-1"},
		}

		repo.On("GetByTeraluxIDPaginated", teraID, 0, 0).Return(expectedDevices, int64(2), nil).Once()

		res, err := useCase.ListDevices(filter)
		assert.NoError(t, err)
		assert.Equal(t, 2, res.Total)
		assert.Len(t, res.Devices, 2)
		repo.AssertExpectations(t)
	})

	// 2. Get All Devices (Success - Empty)
	t.Run("Get All Devices (Success - Empty)", func(t *testing.T) {
		repo := new(MockDeviceRepository)
		useCase := NewGetAllDevicesUseCase(repo)
		teraID := "tx-999"
		filter := &dtos.DeviceFilterDTO{TeraluxID: &teraID}
		repo.On("GetByTeraluxIDPaginated", teraID, 0, 0).Return([]entities.Device{}, int64(0), nil).Once()

		res, err := useCase.ListDevices(filter)
		assert.NoError(t, err)
		assert.Equal(t, 0, res.Total)
		assert.Len(t, res.Devices, 0)
		repo.AssertExpectations(t)
	})
}
