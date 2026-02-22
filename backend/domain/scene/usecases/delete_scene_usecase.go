package usecases

import (
	"teralux_app/domain/scene/repositories"
)

type DeleteSceneUseCase struct {
	repo repositories.ISceneRepository
}

func NewDeleteSceneUseCase(repo repositories.ISceneRepository) *DeleteSceneUseCase {
	return &DeleteSceneUseCase{repo: repo}
}

func (u *DeleteSceneUseCase) DeleteScene(teraluxID, id string) error {
	return u.repo.Delete(teraluxID, id)
}
