package usecases

import (
	"sensio/domain/scene/entities"
	"sensio/domain/scene/repositories"
)

type UpdateSceneUseCase struct {
	repo repositories.ISceneRepository
}

func NewUpdateSceneUseCase(repo repositories.ISceneRepository) *UpdateSceneUseCase {
	return &UpdateSceneUseCase{repo: repo}
}

func (u *UpdateSceneUseCase) UpdateScene(terminalID, id string, name string, actions entities.Actions) error {
	scene, err := u.repo.GetByID(terminalID, id)
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
