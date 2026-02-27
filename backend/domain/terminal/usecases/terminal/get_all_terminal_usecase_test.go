package usecases

import (
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllTerminal_UserBehavior(t *testing.T) {
	repo := new(MockTerminalRepository)
	useCase := NewGetAllTerminalUseCase(repo)

	// 1. Get All Terminal (Success - Empty List)
	t.Run("Get All Terminal (Success - Empty List)", func(t *testing.T) {
		filter := &dtos.TerminalFilterDTO{Page: 1, Limit: 10}
		repo.On("GetAllPaginated", 0, 10, (*string)(nil)).Return([]entities.Terminal{}, int64(0), nil).Once()

		res, err := useCase.ListTerminal(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), res.Total)
		assert.Len(t, res.Terminal, 0)
		repo.AssertExpectations(t)
	})

	// 2. Get All Terminal (Success - With Data)
	t.Run("Get All Terminal (Success - With Data)", func(t *testing.T) {
		filter := &dtos.TerminalFilterDTO{Page: 1, Limit: 10}
		expectedTerminal := make([]entities.Terminal, 10)
		for i := 0; i < 10; i++ {
			expectedTerminal[i] = entities.Terminal{ID: "t", Name: "Hub", DeviceTypeID: "1"}
		}

		repo.On("GetAllPaginated", 0, 10, (*string)(nil)).Return(expectedTerminal, int64(15), nil).Once()

		res, err := useCase.ListTerminal(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), res.Total)
		assert.Len(t, res.Terminal, 10)
		assert.Equal(t, "1", res.Terminal[0].DeviceTypeID)
		repo.AssertExpectations(t)
	})

	// 3. Pagination: Limit and Page
	t.Run("Pagination: Limit and Page", func(t *testing.T) {
		filter := &dtos.TerminalFilterDTO{Page: 2, Limit: 5}
		expectedTerminal := make([]entities.Terminal, 5)
		repo.On("GetAllPaginated", 5, 5, (*string)(nil)).Return(expectedTerminal, int64(15), nil).Once()

		res, err := useCase.ListTerminal(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), res.Total)
		assert.Len(t, res.Terminal, 5)
		assert.Equal(t, 2, res.Page)
		repo.AssertExpectations(t)
	})

	// 4. Filter: By Room ID
	t.Run("Filter: By Room ID", func(t *testing.T) {
		roomID := "r1"
		filter := &dtos.TerminalFilterDTO{Page: 1, Limit: 10, RoomID: &roomID}
		expectedTerminal := []entities.Terminal{{ID: "t1", RoomID: "r1"}}

		repo.On("GetAllPaginated", 0, 10, &roomID).Return(expectedTerminal, int64(1), nil).Once()

		res, err := useCase.ListTerminal(filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.Total)
		assert.Len(t, res.Terminal, 1)
		assert.Equal(t, "r1", res.Terminal[0].RoomID)
		repo.AssertExpectations(t)
	})
}
