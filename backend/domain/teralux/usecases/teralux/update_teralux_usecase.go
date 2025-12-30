package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// UpdateTeraluxUseCase handles the business logic for updating a teralux
type UpdateTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewUpdateTeraluxUseCase creates a new instance of UpdateTeraluxUseCase
func NewUpdateTeraluxUseCase(repository *repositories.TeraluxRepository) *UpdateTeraluxUseCase {
	return &UpdateTeraluxUseCase{
		repository: repository,
	}
}

// Execute updates an existing teralux record
func (uc *UpdateTeraluxUseCase) Execute(id string, req *dtos.UpdateTeraluxRequestDTO) error {
	// Check if teralux exists
	teralux, err := uc.repository.GetByID(id)
	if err != nil {
		return err
	}

	// Update fields if provided
	if req.MacAddress != "" {
		teralux.MacAddress = req.MacAddress
	}
	if req.Name != "" {
		teralux.Name = req.Name
	}

	// Save to database
	if err := uc.repository.Update(teralux); err != nil {
		return err
	}

	return nil
}
