package usecases

import (
	"errors"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteDeviceUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceRepository)
	statusRepo := new(MockDeviceStatusRepository)
	teraRepo := new(MockTerminalRepository)
	useCase := NewDeleteDeviceUseCase(repo, statusRepo, teraRepo)

	// 1. Delete Device (Success)
	t.Run("Delete Device (Success)", func(t *testing.T) {
		deviceID := "dev-1"
		terminalID := "tx-1"

		repo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID, TerminalID: terminalID}, nil).Once()
		statusRepo.On("DeleteByDeviceID", deviceID).Return(nil).Once()
		repo.On("Delete", deviceID).Return(nil).Once()
		teraRepo.On("InvalidateCache", terminalID).Return(nil).Once()

		err := useCase.DeleteDevice(deviceID)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
		statusRepo.AssertExpectations(t)
		teraRepo.AssertExpectations(t)
	})

	// 2. Delete Device (Not Found)
	t.Run("Delete Device (Not Found)", func(t *testing.T) {
		repo.On("GetByID", "dev-999").Return(nil, errors.New("record not found")).Once()

		err := useCase.DeleteDevice("dev-999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Device not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid ID Format
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		err := useCase.DeleteDevice("INVALID")
		assert.Error(t, err)
		assert.Equal(t, "Invalid ID format", err.Error())
	})
}
