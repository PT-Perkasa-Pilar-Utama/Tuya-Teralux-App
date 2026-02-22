package usecases

import (
	"teralux_app/domain/scene/entities"
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

func (m *MockSceneRepository) GetByID(teraluxID, id string) (*entities.Scene, error) {
	args := m.Called(teraluxID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Scene), args.Error(1)
}

func (m *MockSceneRepository) GetAll(teraluxID string) ([]entities.Scene, error) {
	args := m.Called(teraluxID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Scene), args.Error(1)
}

func (m *MockSceneRepository) Delete(teraluxID, id string) error {
	args := m.Called(teraluxID, id)
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

	teraluxID := "test-teralux"
	sceneName := "Test Scene"
	sceneID := "test-scene-id"
	actions := entities.Actions{{DeviceID: "dev1", Code: "on", Value: 1}}

	t.Run("AddScene", func(t *testing.T) {
		usecase := NewAddSceneUseCase(repo)
		repo.On("Save", mock.MatchedBy(func(s *entities.Scene) bool {
			return s.TeraluxID == teraluxID && s.Name == sceneName
		})).Return(nil)

		id, err := usecase.AddScene(teraluxID, sceneName, actions)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		repo.AssertExpectations(t)
	})

	t.Run("GetAllScenes", func(t *testing.T) {
		usecase := NewGetAllScenesUseCase(repo)
		expectedScenes := []entities.Scene{{ID: sceneID, TeraluxID: teraluxID, Name: sceneName}}
		repo.On("GetAll", teraluxID).Return(expectedScenes, nil)

		scenes, err := usecase.ListScenes(teraluxID)
		assert.NoError(t, err)
		assert.NotEmpty(t, scenes)
		assert.Equal(t, sceneName, scenes[0].Name)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateScene", func(t *testing.T) {
		usecase := NewUpdateSceneUseCase(repo)
		repo.On("GetByID", teraluxID, sceneID).Return(&entities.Scene{ID: sceneID, Name: sceneName}, nil)
		repo.On("Save", mock.MatchedBy(func(s *entities.Scene) bool {
			return s.ID == sceneID && s.Name == "Updated Scene"
		})).Return(nil)

		err := usecase.UpdateScene(teraluxID, sceneID, "Updated Scene", nil)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("DeleteScene", func(t *testing.T) {
		usecase := NewDeleteSceneUseCase(repo)
		repo.On("Delete", teraluxID, sceneID).Return(nil)

		err := usecase.DeleteScene(teraluxID, sceneID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
