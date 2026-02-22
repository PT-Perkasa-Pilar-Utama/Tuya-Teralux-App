package usecases

import (
	"errors"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTeraluxByID_UserBehavior(t *testing.T) {
	repo := new(MockTeraluxRepository)
	devRepo := new(MockDeviceRepository)
	useCase := NewGetTeraluxByIDUseCase(repo, devRepo)

	// 1. Get Teralux By ID (Success)
	t.Run("Get Teralux By ID (Success)", func(t *testing.T) {
		id := "t1"
		repo.On("GetByID", id).Return(&entities.Teralux{ID: id, Name: "Living Room"}, nil).Once()
		devRepo.On("GetByTeraluxID", id).Return([]entities.Device{{ID: "d1", Name: "Light"}}, nil).Once()

		res, err := useCase.GetTeraluxByID(id)
		assert.NoError(t, err)
		assert.Equal(t, id, res.Teralux.ID)
		
		repo.AssertExpectations(t)
		devRepo.AssertExpectations(t)
	})

	// 2. Get Teralux By ID (Not Found)
	t.Run("Get Teralux By ID (Not Found)", func(t *testing.T) {
		repo.On("GetByID", "unknown-id").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetTeraluxByID("unknown-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Teralux not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid ID Format
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		_, err := useCase.GetTeraluxByID("INVALID-FORMAT")
		assert.Error(t, err)
		assert.Equal(t, "Invalid ID format", err.Error())
	})
}
