package usecases

import (
	"errors"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteTerminal_UserBehavior(t *testing.T) {
	repo := new(MockTerminalRepository)
	useCase := NewDeleteTerminalUseCase(repo)

	// 1. Delete Terminal (Success Condition)
	t.Run("Delete Terminal (Success Condition)", func(t *testing.T) {
		id := "tx-1"
		repo.On("GetByID", id).Return(&entities.Terminal{ID: id}, nil).Once()
		repo.On("Delete", id).Return(nil).Once()
		repo.On("InvalidateCache", id).Return(nil).Once()

		err := useCase.DeleteTerminal(id)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// 2. Delete Terminal (Not Found)
	t.Run("Delete Terminal (Not Found)", func(t *testing.T) {
		repo.On("GetByID", "tx-999").Return(nil, errors.New("record not found")).Once()

		err := useCase.DeleteTerminal("tx-999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terminal not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid ID Format
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		err := useCase.DeleteTerminal("INVALID-UUID")
		assert.Error(t, err)
		assert.Equal(t, "Invalid ID format", err.Error())
	})
}
