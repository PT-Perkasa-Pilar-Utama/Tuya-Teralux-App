package usecases

import (
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"
)

type GetAllScenesUseCase struct {
	repo *repositories.SceneRepository
}

func NewGetAllScenesUseCase(repo *repositories.SceneRepository) *GetAllScenesUseCase {
	return &GetAllScenesUseCase{repo: repo}
}

func (u *GetAllScenesUseCase) ListScenes(teraluxID string) ([]entities.Scene, error) {
	return u.repo.GetAll(teraluxID)
}
