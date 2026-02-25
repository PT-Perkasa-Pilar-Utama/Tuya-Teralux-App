package usecases

import (
	"errors"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTerminalByMAC_UserBehavior(t *testing.T) {
	repo := new(MockTerminalRepository)
	useCase := NewGetTerminalByMACUseCase(repo)

	// 1. Get Terminal By MAC (Success)
	t.Run("Get Terminal By MAC (Success)", func(t *testing.T) {
		mac := "AA:BB:CC:11:22:33"
		repo.On("GetByMacAddress", mac).Return(&entities.Terminal{ID: "t1", MacAddress: mac}, nil).Once()

		res, err := useCase.GetTerminalByMAC(mac)
		assert.NoError(t, err)
		assert.Equal(t, mac, res.Terminal.MacAddress)
		repo.AssertExpectations(t)
	})

	// 2. Get Terminal By MAC (Not Found)
	t.Run("Get Terminal By MAC (Not Found)", func(t *testing.T) {
		repo.On("GetByMacAddress", "AA:BB:CC:DD:EE:FF").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetTerminalByMAC("AA:BB:CC:DD:EE:FF")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terminal not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid MAC Format
	t.Run("Validation: Invalid MAC Format", func(t *testing.T) {
		_, err := useCase.GetTerminalByMAC("INVALID-MAC")
		assert.Error(t, err)
		assert.Equal(t, "invalid mac address or device id format", err.Error())
	})
}
