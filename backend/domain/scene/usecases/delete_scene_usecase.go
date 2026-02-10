package usecases

import "teralux_app/domain/scene/repositories"

type DeleteSceneUseCase struct {
	repo *repositories.SceneRepository
}

func NewDeleteSceneUseCase(repo *repositories.SceneRepository) *DeleteSceneUseCase {
	return &DeleteSceneUseCase{repo: repo}
}

func (uc *DeleteSceneUseCase) Execute(id string) error {
	return uc.repo.Delete(id)
}
