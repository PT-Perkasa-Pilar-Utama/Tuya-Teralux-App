package usecases

import (
	"errors"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTeraluxByMAC_UserBehavior(t *testing.T) {
	repo := new(MockTeraluxRepository)
	useCase := NewGetTeraluxByMACUseCase(repo)

	// 1. Get Teralux By MAC (Success)
	t.Run("Get Teralux By MAC (Success)", func(t *testing.T) {
		mac := "AA:BB:CC:11:22:33"
		repo.On("GetByMacAddress", mac).Return(&entities.Teralux{ID: "t1", MacAddress: mac}, nil).Once()

		res, err := useCase.GetTeraluxByMAC(mac)
		assert.NoError(t, err)
		assert.Equal(t, mac, res.Teralux.MacAddress)
		repo.AssertExpectations(t)
	})

	// 2. Get Teralux By MAC (Not Found)
	t.Run("Get Teralux By MAC (Not Found)", func(t *testing.T) {
		repo.On("GetByMacAddress", "AA:BB:CC:DD:EE:FF").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetTeraluxByMAC("AA:BB:CC:DD:EE:FF")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Teralux not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid MAC Format
	t.Run("Validation: Invalid MAC Format", func(t *testing.T) {
		_, err := useCase.GetTeraluxByMAC("INVALID-MAC")
		assert.Error(t, err)
		assert.Equal(t, "invalid mac address or device id format", err.Error())
	})
}
