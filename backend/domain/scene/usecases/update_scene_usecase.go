package usecases

import (
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"
)

type UpdateSceneUseCase struct {
	repo *repositories.SceneRepository
}

func NewUpdateSceneUseCase(repo *repositories.SceneRepository) *UpdateSceneUseCase {
	return &UpdateSceneUseCase{repo: repo}
}

func (u *UpdateSceneUseCase) UpdateScene(teraluxID, id string, name string, actions entities.Actions) error {
	scene, err := u.repo.GetByID(teraluxID, id)
	if err != nil {
		return err
	}

	if name != "" {
		scene.Name = name
	}
	if actions != nil {
		scene.Actions = actions
	}

	return u.repo.Save(scene)
}
