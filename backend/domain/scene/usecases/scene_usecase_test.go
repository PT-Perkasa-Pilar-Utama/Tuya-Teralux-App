package usecases

import (
	"sensio/domain/scene/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSceneRepository struct {
	mock.Mock
}

func (m *MockSceneRepository) Save(scene *entities.Scene) error {
	args := m.Called(scene)
	return args.Error(0)
}

func (m *MockSceneRepository) GetByID(terminalID, id string) (*entities.Scene, error) {
	args := m.Called(terminalID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Scene), args.Error(1)
}

func (m *MockSceneRepository) GetAll(terminalID string) ([]entities.Scene, error) {
	args := m.Called(terminalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Scene), args.Error(1)
}

func (m *MockSceneRepository) Delete(terminalID, id string) error {
	args := m.Called(terminalID, id)
	return args.Error(0)
}

func (m *MockSceneRepository) GetAllGrouped() (map[string][]entities.Scene, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]entities.Scene), args.Error(1)
}

func TestSceneUsecases(t *testing.T) {
	repo := new(MockSceneRepository)

	terminalID := "test-terminal"
	sceneName := "Test Scene"
	sceneID := "test-scene-id"
	actions := entities.Actions{{DeviceID: "dev1", Code: "on", Value: 1}}

	t.Run("AddScene", func(t *testing.T) {
		usecase := NewAddSceneUseCase(repo)
		repo.On("Save", mock.MatchedBy(func(s *entities.Scene) bool {
			return s.TerminalID == terminalID && s.Name == sceneName
		})).Return(nil)

		id, err := usecase.AddScene(terminalID, sceneName, actions)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		repo.AssertExpectations(t)
	})

	t.Run("GetAllScenes", func(t *testing.T) {
		usecase := NewGetAllScenesUseCase(repo)
		expectedScenes := []entities.Scene{{ID: sceneID, TerminalID: terminalID, Name: sceneName}}
		repo.On("GetAll", terminalID).Return(expectedScenes, nil)

		scenes, err := usecase.ListScenes(terminalID)
		assert.NoError(t, err)
		assert.NotEmpty(t, scenes)
		assert.Equal(t, sceneName, scenes[0].Name)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateScene", func(t *testing.T) {
		usecase := NewUpdateSceneUseCase(repo)
		repo.On("GetByID", terminalID, sceneID).Return(&entities.Scene{ID: sceneID, Name: sceneName}, nil)
		repo.On("Save", mock.MatchedBy(func(s *entities.Scene) bool {
			return s.ID == sceneID && s.Name == "Updated Scene"
		})).Return(nil)

		err := usecase.UpdateScene(terminalID, sceneID, "Updated Scene", nil)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("DeleteScene", func(t *testing.T) {
		usecase := NewDeleteSceneUseCase(repo)
		repo.On("Delete", terminalID, sceneID).Return(nil)

		err := usecase.DeleteScene(terminalID, sceneID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
