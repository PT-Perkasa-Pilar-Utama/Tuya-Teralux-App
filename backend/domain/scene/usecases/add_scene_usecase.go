package usecases

import (
	"teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"

	"github.com/google/uuid"
)

type AddSceneUseCase struct {
	repo *repositories.SceneRepository
}

func NewAddSceneUseCase(repo *repositories.SceneRepository) *AddSceneUseCase {
	return &AddSceneUseCase{repo: repo}
}

func (uc *AddSceneUseCase) Execute(req *dtos.CreateSceneRequestDTO) (string, error) {
	id := uuid.New().String()
	
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

	if err := uc.repo.Save(scene); err != nil {
		return "", err
	}

	return id, nil
}
