package usecases

import (
	"errors"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTerminalByID_UserBehavior(t *testing.T) {
	repo := new(MockTerminalRepository)
	devRepo := new(MockDeviceRepository)
	useCase := NewGetTerminalByIDUseCase(repo, devRepo)

	// 1. Get Terminal By ID (Success)
	t.Run("Get Terminal By ID (Success)", func(t *testing.T) {
		id := "t1"
		repo.On("GetByID", id).Return(&entities.Terminal{ID: id, Name: "Living Room", DeviceTypeID: "1"}, nil).Once()
		devRepo.On("GetByTerminalID", id).Return([]entities.Device{{ID: "d1", Name: "Light"}}, nil).Once()

		res, err := useCase.GetTerminalByID(id)
		assert.NoError(t, err)
		assert.Equal(t, id, res.Terminal.ID)
		assert.Equal(t, "1", res.Terminal.DeviceTypeID)

		repo.AssertExpectations(t)
		devRepo.AssertExpectations(t)
	})

	// 2. Get Terminal By ID (Not Found)
	t.Run("Get Terminal By ID (Not Found)", func(t *testing.T) {
		repo.On("GetByID", "unknown-id").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetTerminalByID("unknown-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terminal not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid ID Format
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		_, err := useCase.GetTerminalByID("INVALID-FORMAT")
		assert.Error(t, err)
		assert.Equal(t, "Invalid ID format", err.Error())
	})
}
