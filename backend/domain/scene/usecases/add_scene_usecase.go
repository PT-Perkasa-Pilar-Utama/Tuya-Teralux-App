package usecases

import (
	"sensio/domain/scene/entities"
	"sensio/domain/scene/repositories"

	"github.com/google/uuid"
)

type AddSceneUseCase struct {
	repo repositories.ISceneRepository
}

func NewAddSceneUseCase(repo repositories.ISceneRepository) *AddSceneUseCase {
	return &AddSceneUseCase{repo: repo}
}

func (u *AddSceneUseCase) AddScene(terminalID string, name string, actions entities.Actions) (string, error) {
	scene := &entities.Scene{
		ID:         uuid.New().String(),
		TerminalID: terminalID,
		Name:       name,
		Actions:    actions,
	}

	if err := u.repo.Save(scene); err != nil {
		return "", err
	}

	return scene.ID, nil
}
