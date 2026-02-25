package usecases

import (
	"errors"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTeralux_UserBehavior(t *testing.T) {
	// 1. Update Teralux (Success - Name Only)
	t.Run("Update Teralux (Success - Name Only)", func(t *testing.T) {
		repo := new(MockTeraluxRepository)
		useCase := NewUpdateTeraluxUseCase(repo)
		id := "t1"
		newName := "New Name"
		req := &dtos.UpdateTeraluxRequestDTO{
			Name: &newName,
		}
		repo.On("GetByID", id).Return(&entities.Teralux{ID: id, Name: "Old Name"}, nil).Once()
		repo.On("Update", mock.MatchedBy(func(teralux *entities.Teralux) bool {
			return teralux.Name == newName
		})).Return(nil).Once()
		repo.On("InvalidateCache", id).Return(nil).Once()

		err := useCase.UpdateTeralux(id, req)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// 2. Update Teralux (Success - Move Room)
	t.Run("Update Teralux (Success - Move Room)", func(t *testing.T) {
		repo := new(MockTeraluxRepository)
		useCase := NewUpdateTeraluxUseCase(repo)
		id := "t1"
		newRoom := "room-2"
		req := &dtos.UpdateTeraluxRequestDTO{
			RoomID: &newRoom,
		}
		repo.On("GetByID", id).Return(&entities.Teralux{ID: id, RoomID: "r1"}, nil).Once()
		repo.On("Update", mock.MatchedBy(func(teralux *entities.Teralux) bool {
			return teralux.RoomID == newRoom
		})).Return(nil).Once()
		repo.On("InvalidateCache", id).Return(nil).Once()

		err := useCase.UpdateTeralux(id, req)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// 3. Update Teralux (Not Found)
	t.Run("Update Teralux (Not Found)", func(t *testing.T) {
		repo := new(MockTeraluxRepository)
		useCase := NewUpdateTeraluxUseCase(repo)
		name := "Hack"
		req := &dtos.UpdateTeraluxRequestDTO{Name: &name}
		repo.On("GetByID", "unknown").Return(nil, errors.New("record not found")).Once()

		err := useCase.UpdateTeralux("unknown", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Teralux not found")
		repo.AssertExpectations(t)
	})

	// 5. Validation: Empty Name (If Present)
	t.Run("Validation: Empty Name (If Present)", func(t *testing.T) {
		repo := new(MockTeraluxRepository)
		useCase := NewUpdateTeraluxUseCase(repo)
		id := "t1"
		emptyName := ""
		req := &dtos.UpdateTeraluxRequestDTO{Name: &emptyName}
		repo.On("GetByID", id).Return(&entities.Teralux{ID: id}, nil).Once()

		err := useCase.UpdateTeralux(id, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
		repo.AssertExpectations(t)
	})

	// 6. Conflict: Update to Duplicate MAC
	t.Run("Conflict: Update to Duplicate MAC", func(t *testing.T) {
		repo := new(MockTeraluxRepository)
		useCase := NewUpdateTeraluxUseCase(repo)
		id := "t1"
		duplicateMac := "MAC-2"
		req := &dtos.UpdateTeraluxRequestDTO{MacAddress: &duplicateMac}

		repo.On("GetByID", id).Return(&entities.Teralux{ID: id}, nil).Once()
		repo.On("GetByMacAddress", duplicateMac).Return(&entities.Teralux{ID: "t2"}, nil).Once()

		err := useCase.UpdateTeralux(id, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mac Address already in use")
		repo.AssertExpectations(t)
	})
}
