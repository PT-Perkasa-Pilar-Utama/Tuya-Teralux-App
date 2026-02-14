package usecases

import (
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&entities.Scene{})
	return db
}

func TestSceneUsecases(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewSceneRepository(db)

	teraluxID := "test-teralux"
	sceneName := "Test Scene"
	actions := entities.Actions{{DeviceID: "dev1", Code: "on", Value: 1}}

	t.Run("AddScene", func(t *testing.T) {
		usecase := NewAddSceneUseCase(repo)
		id, err := usecase.AddScene(teraluxID, sceneName, actions)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)

		// Verify in DB
		scene, err := repo.GetByID(teraluxID, id)
		assert.NoError(t, err)
		assert.Equal(t, sceneName, scene.Name)
	})

	t.Run("GetAllScenes", func(t *testing.T) {
		usecase := NewGetAllScenesUseCase(repo)
		scenes, err := usecase.ListScenes(teraluxID)
		assert.NoError(t, err)
		assert.NotEmpty(t, scenes)
		assert.Equal(t, sceneName, scenes[0].Name)
	})

	t.Run("UpdateScene", func(t *testing.T) {
		// Get existing ID
		scenes, _ := repo.GetAll(teraluxID)
		id := scenes[0].ID

		usecase := NewUpdateSceneUseCase(repo)
		newName := "Updated Scene"
		err := usecase.UpdateScene(teraluxID, id, newName, nil)
		assert.NoError(t, err)

		// Verify
		scene, _ := repo.GetByID(teraluxID, id)
		assert.Equal(t, newName, scene.Name)
	})

	t.Run("DeleteScene", func(t *testing.T) {
		scenes, _ := repo.GetAll(teraluxID)
		id := scenes[0].ID

		usecase := NewDeleteSceneUseCase(repo)
		err := usecase.DeleteScene(teraluxID, id)
		assert.NoError(t, err)

		// Verify
		_, err = repo.GetByID(teraluxID, id)
		assert.Error(t, err)
	})
}
