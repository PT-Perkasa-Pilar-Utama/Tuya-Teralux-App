package usecases

import (
	"errors"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTerminal_UserBehavior(t *testing.T) {
	// 1. Update Terminal (Success - Name Only)
	t.Run("Update Terminal (Success - Name Only)", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewUpdateTerminalUseCase(repo)
		id := "t1"
		newName := "New Name"
		req := &dtos.UpdateTerminalRequestDTO{
			Name: &newName,
		}
		repo.On("GetByID", id).Return(&entities.Terminal{ID: id, Name: "Old Name"}, nil).Once()
		repo.On("Update", mock.MatchedBy(func(terminal *entities.Terminal) bool {
			return terminal.Name == newName
		})).Return(nil).Once()
		repo.On("InvalidateCache", id).Return(nil).Once()

		err := useCase.UpdateTerminal(id, req)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// 2. Update Terminal (Success - Move Room)
	t.Run("Update Terminal (Success - Move Room)", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewUpdateTerminalUseCase(repo)
		id := "t1"
		newRoom := "room-2"
		req := &dtos.UpdateTerminalRequestDTO{
			RoomID: &newRoom,
		}
		repo.On("GetByID", id).Return(&entities.Terminal{ID: id, RoomID: "r1"}, nil).Once()
		repo.On("Update", mock.MatchedBy(func(terminal *entities.Terminal) bool {
			return terminal.RoomID == newRoom
		})).Return(nil).Once()
		repo.On("InvalidateCache", id).Return(nil).Once()

		err := useCase.UpdateTerminal(id, req)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// 3. Update Terminal (Not Found)
	t.Run("Update Terminal (Not Found)", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewUpdateTerminalUseCase(repo)
		name := "Hack"
		req := &dtos.UpdateTerminalRequestDTO{Name: &name}
		repo.On("GetByID", "unknown").Return(nil, errors.New("record not found")).Once()

		err := useCase.UpdateTerminal("unknown", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terminal not found")
		repo.AssertExpectations(t)
	})

	// 5. Validation: Empty Name (If Present)
	t.Run("Validation: Empty Name (If Present)", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewUpdateTerminalUseCase(repo)
		id := "t1"
		emptyName := ""
		req := &dtos.UpdateTerminalRequestDTO{Name: &emptyName}
		repo.On("GetByID", id).Return(&entities.Terminal{ID: id}, nil).Once()

		err := useCase.UpdateTerminal(id, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
		repo.AssertExpectations(t)
	})

	// 6. Conflict: Update to Duplicate MAC
	t.Run("Conflict: Update to Duplicate MAC", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewUpdateTerminalUseCase(repo)
		id := "t1"
		duplicateMac := "MAC-2"
		req := &dtos.UpdateTerminalRequestDTO{MacAddress: &duplicateMac}

		repo.On("GetByID", id).Return(&entities.Terminal{ID: id}, nil).Once()
		repo.On("GetByMacAddress", duplicateMac).Return(&entities.Terminal{ID: "t2"}, nil).Once()

		err := useCase.UpdateTerminal(id, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mac Address already in use")
		repo.AssertExpectations(t)
	})
}
