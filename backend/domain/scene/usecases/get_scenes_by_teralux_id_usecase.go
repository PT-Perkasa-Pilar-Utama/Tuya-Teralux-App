package usecases

import (
	"sensio/domain/scene/entities"
	"sensio/domain/scene/repositories"
)

type GetAllScenesUseCase struct {
	repo repositories.ISceneRepository
}

func NewGetAllScenesUseCase(repo repositories.ISceneRepository) *GetAllScenesUseCase {
	return &GetAllScenesUseCase{repo: repo}
}

func (u *GetAllScenesUseCase) ListScenes(terminalID string) ([]entities.Scene, error) {
	return u.repo.GetAll(terminalID)
}
