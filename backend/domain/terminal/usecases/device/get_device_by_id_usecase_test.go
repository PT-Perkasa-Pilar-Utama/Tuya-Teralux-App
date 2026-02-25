package usecases

import (
	"errors"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDeviceByIDUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceRepository)
	useCase := NewGetDeviceByIDUseCase(repo)

	// 1. Get Device By ID (Success)
	t.Run("Get Device By ID (Success)", func(t *testing.T) {
		id := "d1"
		repo.On("GetByID", id).Return(&entities.Device{ID: id, Name: "Light 1"}, nil).Once()

		res, err := useCase.GetDeviceByID(id)
		assert.NoError(t, err)
		assert.Equal(t, id, res.Device.ID)
		repo.AssertExpectations(t)
	})

	// 2. Get Device By ID (Not Found)
	t.Run("Get Device By ID (Not Found)", func(t *testing.T) {
		repo.On("GetByID", "unknown").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetDeviceByID("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Device not found")
		repo.AssertExpectations(t)
	})
}
