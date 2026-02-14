package usecases

import (
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"

	"github.com/google/uuid"
)

type AddSceneUseCase struct {
	repo *repositories.SceneRepository
}

func NewAddSceneUseCase(repo *repositories.SceneRepository) *AddSceneUseCase {
	return &AddSceneUseCase{repo: repo}
}

func (u *AddSceneUseCase) AddScene(teraluxID string, name string, actions entities.Actions) (string, error) {
	scene := &entities.Scene{
		ID:        uuid.New().String(),
		TeraluxID: teraluxID,
		Name:      name,
		Actions:   actions,
	}

	if err := u.repo.Save(scene); err != nil {
		return "", err
	}

	return scene.ID, nil
}
