package usecases

import (
	"errors"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateDeviceUseCase_UserBehavior(t *testing.T) {
	// 1. Update Device (Success)
	t.Run("Update Device (Success)", func(t *testing.T) {
		repo := new(MockDeviceRepository)
		teraRepo := new(MockTeraluxRepository)
		useCase := NewUpdateDeviceUseCase(repo, teraRepo)
		id := "d1"
		newName := "Updated Name"
		req := &dtos.UpdateDeviceRequestDTO{Name: &newName}
		
		repo.On("GetByID", id).Return(&entities.Device{ID: id, Name: "Old Name", TeraluxID: "tx-1"}, nil).Once()
		repo.On("Update", mock.MatchedBy(func(d *entities.Device) bool {
			return d.ID == id && d.Name == newName
		})).Return(nil).Once()
		teraRepo.On("InvalidateCache", "tx-1").Return(nil).Once()

		err := useCase.UpdateDevice(id, req)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		teraRepo.AssertExpectations(t)
	})

	// 2. Update Device (Not Found)
	t.Run("Update Device (Not Found)", func(t *testing.T) {
		repo := new(MockDeviceRepository)
		teraRepo := new(MockTeraluxRepository)
		useCase := NewUpdateDeviceUseCase(repo, teraRepo)
		name := "Ghost"
		req := &dtos.UpdateDeviceRequestDTO{Name: &name}
		repo.On("GetByID", "unknown").Return(nil, errors.New("record not found")).Once()

		err := useCase.UpdateDevice("unknown", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Device not found")
		repo.AssertExpectations(t)
	})
}
