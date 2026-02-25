package usecases

import (
	"sensio/domain/scene/repositories"
)

type DeleteSceneUseCase struct {
	repo repositories.ISceneRepository
}

func NewDeleteSceneUseCase(repo repositories.ISceneRepository) *DeleteSceneUseCase {
	return &DeleteSceneUseCase{repo: repo}
}

func (u *DeleteSceneUseCase) DeleteScene(terminalID, id string) error {
	return u.repo.Delete(terminalID, id)
}
