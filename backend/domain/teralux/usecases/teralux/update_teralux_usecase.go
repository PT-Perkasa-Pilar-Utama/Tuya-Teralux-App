package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// UpdateTeraluxUseCase handles updating an existing teralux
type UpdateTeraluxUseCase struct {
	repository TeraluxRepository
}

// NewUpdateTeraluxUseCase creates a new instance of UpdateTeraluxUseCase
func NewUpdateTeraluxUseCase(repository TeraluxRepository) *UpdateTeraluxUseCase {
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
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.MacAddress != "" {
		item.MacAddress = req.MacAddress
	}

	// Save changes
	return uc.repository.Update(item)
}
