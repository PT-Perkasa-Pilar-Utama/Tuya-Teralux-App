package usecases

import (
	"errors"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteTeralux_UserBehavior(t *testing.T) {
	repo := new(MockTeraluxRepository)
	useCase := NewDeleteTeraluxUseCase(repo)

	// 1. Delete Teralux (Success Condition)
	t.Run("Delete Teralux (Success Condition)", func(t *testing.T) {
		id := "tx-1"
		repo.On("GetByID", id).Return(&entities.Teralux{ID: id}, nil).Once()
		repo.On("Delete", id).Return(nil).Once()
		repo.On("InvalidateCache", id).Return(nil).Once()

		err := useCase.DeleteTeralux(id)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// 2. Delete Teralux (Not Found)
	t.Run("Delete Teralux (Not Found)", func(t *testing.T) {
		repo.On("GetByID", "tx-999").Return(nil, errors.New("record not found")).Once()

		err := useCase.DeleteTeralux("tx-999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Teralux not found")
		repo.AssertExpectations(t)
	})

	// 3. Validation: Invalid ID Format
	t.Run("Validation: Invalid ID Format", func(t *testing.T) {
		err := useCase.DeleteTeralux("INVALID-UUID")
		assert.Error(t, err)
		assert.Equal(t, "Invalid ID format", err.Error())
	})
}
