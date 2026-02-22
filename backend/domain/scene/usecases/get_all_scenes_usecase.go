package usecases

import (
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"
)

type GetAllGroupedScenesUseCase struct {
	repo *repositories.SceneRepository
}

func NewGetAllGroupedScenesUseCase(repo *repositories.SceneRepository) *GetAllGroupedScenesUseCase {
	return &GetAllGroupedScenesUseCase{repo: repo}
}

func (u *GetAllGroupedScenesUseCase) ListAllGrouped() (map[string][]entities.Scene, error) {
	return u.repo.GetAllGrouped()
}
