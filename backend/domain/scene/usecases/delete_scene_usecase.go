package usecases

import (
	"teralux_app/domain/scene/repositories"
)

type DeleteSceneUseCase struct {
	repo *repositories.SceneRepository
}

func NewDeleteSceneUseCase(repo *repositories.SceneRepository) *DeleteSceneUseCase {
	return &DeleteSceneUseCase{repo: repo}
}

func (u *DeleteSceneUseCase) DeleteScene(teraluxID, id string) error {
	return u.repo.Delete(teraluxID, id)
}
