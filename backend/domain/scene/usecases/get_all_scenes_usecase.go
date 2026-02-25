package usecases

import (
	"sensio/domain/scene/entities"
	"sensio/domain/scene/repositories"
)

type GetAllGroupedScenesUseCase struct {
	repo repositories.ISceneRepository
}

func NewGetAllGroupedScenesUseCase(repo repositories.ISceneRepository) *GetAllGroupedScenesUseCase {
	return &GetAllGroupedScenesUseCase{repo: repo}
}

func (u *GetAllGroupedScenesUseCase) ListAllGrouped() (map[string][]entities.Scene, error) {
	return u.repo.GetAllGrouped()
}
