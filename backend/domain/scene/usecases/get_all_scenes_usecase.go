package usecases

import (
	"teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/repositories"
)

type GetAllScenesUseCase struct {
	repo *repositories.SceneRepository
}

func NewGetAllScenesUseCase(repo *repositories.SceneRepository) *GetAllScenesUseCase {
	return &GetAllScenesUseCase{repo: repo}
}

func (uc *GetAllScenesUseCase) Execute() ([]dtos.SceneListResponseDTO, error) {
	scenes, err := uc.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var res []dtos.SceneListResponseDTO
	for _, s := range scenes {
		res = append(res, dtos.SceneListResponseDTO{
			ID:   s.ID,
			Name: s.Name,
		})
	}
	return res, nil
}
