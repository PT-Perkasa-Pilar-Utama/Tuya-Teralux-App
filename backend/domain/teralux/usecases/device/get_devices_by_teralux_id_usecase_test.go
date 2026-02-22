package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDevicesByTeraluxIDUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceRepository)
	teraRepo := new(MockTeraluxRepository)
	useCase := NewGetDevicesByTeraluxIDUseCase(repo, teraRepo)

	// 1. Get Devices By Teralux ID (Success)
	t.Run("Get Devices By Teralux ID (Success)", func(t *testing.T) {
		teraID := "tx-1"
		teraRepo.On("GetByID", teraID).Return(&entities.Teralux{ID: teraID}, nil).Once()
		repo.On("GetByTeraluxIDPaginated", teraID, 0, 10).Return([]entities.Device{
			{ID: "d1", Name: "D1", TeraluxID: teraID},
			{ID: "d2", Name: "D2", TeraluxID: teraID},
		}, int64(2), nil).Once()

		res, err := useCase.ListDevicesByTeraluxID(teraID, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 2, res.Total)
		assert.Len(t, res.Devices, 2)
		teraRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	// 2. Get Devices By Teralux ID (Not Found)
	t.Run("Get Devices By Teralux ID (Not Found)", func(t *testing.T) {
		teraRepo.On("GetByID", "tx-999").Return(nil, assert.AnError).Once()

		_, err := useCase.ListDevicesByTeraluxID("tx-999", 1, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Teralux hub not found")
		teraRepo.AssertExpectations(t)
	})
}
