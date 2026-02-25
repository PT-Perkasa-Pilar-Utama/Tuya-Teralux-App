package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllTeralux_UserBehavior(t *testing.T) {
	repo := new(MockTeraluxRepository)
	useCase := NewGetAllTeraluxUseCase(repo)

	// 1. Get All Teralux (Success - Empty List)
	t.Run("Get All Teralux (Success - Empty List)", func(t *testing.T) {
		filter := &dtos.TeraluxFilterDTO{Page: 1, Limit: 10}
		repo.On("GetAllPaginated", 0, 10, (*string)(nil)).Return([]entities.Teralux{}, int64(0), nil).Once()

		res, err := useCase.ListTeralux(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), res.Total)
		assert.Len(t, res.Teralux, 0)
		repo.AssertExpectations(t)
	})

	// 2. Get All Teralux (Success - With Data)
	t.Run("Get All Teralux (Success - With Data)", func(t *testing.T) {
		filter := &dtos.TeraluxFilterDTO{Page: 1, Limit: 10}
		expectedTeralux := make([]entities.Teralux, 10)
		for i := 0; i < 10; i++ {
			expectedTeralux[i] = entities.Teralux{ID: "t", Name: "Hub"}
		}

		repo.On("GetAllPaginated", 0, 10, (*string)(nil)).Return(expectedTeralux, int64(15), nil).Once()

		res, err := useCase.ListTeralux(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), res.Total)
		assert.Len(t, res.Teralux, 10)
		repo.AssertExpectations(t)
	})

	// 3. Pagination: Limit and Page
	t.Run("Pagination: Limit and Page", func(t *testing.T) {
		filter := &dtos.TeraluxFilterDTO{Page: 2, Limit: 5}
		expectedTeralux := make([]entities.Teralux, 5)
		repo.On("GetAllPaginated", 5, 5, (*string)(nil)).Return(expectedTeralux, int64(15), nil).Once()

		res, err := useCase.ListTeralux(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), res.Total)
		assert.Len(t, res.Teralux, 5)
		assert.Equal(t, 2, res.Page)
		repo.AssertExpectations(t)
	})

	// 4. Filter: By Room ID
	t.Run("Filter: By Room ID", func(t *testing.T) {
		roomID := "r1"
		filter := &dtos.TeraluxFilterDTO{Page: 1, Limit: 10, RoomID: &roomID}
		expectedTeralux := []entities.Teralux{{ID: "t1", RoomID: "r1"}}

		repo.On("GetAllPaginated", 0, 10, &roomID).Return(expectedTeralux, int64(1), nil).Once()

		res, err := useCase.ListTeralux(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.Total)
		assert.Len(t, res.Teralux, 1)
		assert.Equal(t, "r1", res.Teralux[0].RoomID)
		repo.AssertExpectations(t)
	})
}
