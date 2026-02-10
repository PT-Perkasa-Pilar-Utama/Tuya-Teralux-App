package usecases

import (
	"fmt"
	"teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"
)

type UpdateSceneUseCase struct {
	repo *repositories.SceneRepository
}

func NewUpdateSceneUseCase(repo *repositories.SceneRepository) *UpdateSceneUseCase {
	return &UpdateSceneUseCase{repo: repo}
}

func (uc *UpdateSceneUseCase) Execute(id string, req *dtos.UpdateSceneRequestDTO) error {
	// Check exist
	_, err := uc.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("scene not found: %w", err)
	}

	var actions []entities.Action
	for _, a := range req.Actions {
		actions = append(actions, entities.Action{
			DeviceID: a.DeviceID,
			Code:     a.Code,
			RemoteID: a.RemoteID,
			Topic:    a.Topic,
			Value:    a.Value,
		})
	}

	scene := &entities.Scene{
		ID:      id,
		Name:    req.Name,
		Actions: actions,
	}

	return uc.repo.Save(scene)
}
