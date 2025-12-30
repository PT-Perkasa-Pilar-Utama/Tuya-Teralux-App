package usecases

import (
	"teralux_app/domain/teralux/repositories"
)

// DeleteTeraluxUseCase handles the business logic for deleting a teralux
type DeleteTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewDeleteTeraluxUseCase creates a new instance of DeleteTeraluxUseCase
func NewDeleteTeraluxUseCase(repository *repositories.TeraluxRepository) *DeleteTeraluxUseCase {
	return &DeleteTeraluxUseCase{
		repository: repository,
	}
}

// Execute soft deletes a teralux record by ID
func (uc *DeleteTeraluxUseCase) Execute(id string) error {
	// Check if teralux exists before deleting
	_, err := uc.repository.GetByID(id)
	if err != nil {
		return err
	}

	// Soft delete the teralux
	return uc.repository.Delete(id)
}
