package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// UpdateTeraluxUseCase handles updating an existing teralux
type UpdateTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewUpdateTeraluxUseCase creates a new instance of UpdateTeraluxUseCase
func NewUpdateTeraluxUseCase(repository *repositories.TeraluxRepository) *UpdateTeraluxUseCase {
	return &UpdateTeraluxUseCase{
		repository: repository,
	}
}

// Execute updates a teralux
func (uc *UpdateTeraluxUseCase) Execute(id string, req *dtos.UpdateTeraluxRequestDTO) error {
	// First check if exists
	item, err := uc.repository.GetByID(id)
	if err != nil {
		return err
	}

	// Update fields
	if req.RoomID != "" {
		item.RoomID = req.RoomID
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.MacAddress != "" {
		item.MacAddress = req.MacAddress
	}

	// Save changes
	return uc.repository.Update(item)
}
