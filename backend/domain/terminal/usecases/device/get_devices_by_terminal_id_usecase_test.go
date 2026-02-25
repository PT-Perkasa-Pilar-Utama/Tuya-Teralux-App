package usecases

import (
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDevicesByTerminalIDUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceRepository)
	teraRepo := new(MockTerminalRepository)
	useCase := NewGetDevicesByTerminalIDUseCase(repo, teraRepo)

	// 1. Get Devices By Terminal ID (Success)
	t.Run("Get Devices By Terminal ID (Success)", func(t *testing.T) {
		teraID := "tx-1"
		teraRepo.On("GetByID", teraID).Return(&entities.Terminal{ID: teraID}, nil).Once()
		repo.On("GetByTerminalIDPaginated", teraID, 0, 10).Return([]entities.Device{
			{ID: "d1", Name: "D1", TerminalID: teraID},
			{ID: "d2", Name: "D2", TerminalID: teraID},
		}, int64(2), nil).Once()

		res, err := useCase.ListDevicesByTerminalID(teraID, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 2, res.Total)
		assert.Len(t, res.Devices, 2)
		teraRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	// 2. Get Devices By Terminal ID (Not Found)
	t.Run("Get Devices By Terminal ID (Not Found)", func(t *testing.T) {
		teraRepo.On("GetByID", "tx-999").Return(nil, assert.AnError).Once()

		_, err := useCase.ListDevicesByTerminalID("tx-999", 1, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terminal hub not found")
		teraRepo.AssertExpectations(t)
	})
}
