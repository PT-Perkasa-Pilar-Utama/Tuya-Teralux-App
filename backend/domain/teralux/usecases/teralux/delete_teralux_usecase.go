package usecases

import "teralux_app/domain/teralux/repositories"

// DeleteTeraluxUseCase handles deleting a teralux
type DeleteTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewDeleteTeraluxUseCase creates a new instance of DeleteTeraluxUseCase
func NewDeleteTeraluxUseCase(repository *repositories.TeraluxRepository) *DeleteTeraluxUseCase {
	return &DeleteTeraluxUseCase{
		repository: repository,
	}
}

// Execute deletes a teralux by ID
func (uc *DeleteTeraluxUseCase) Execute(id string) error {
	return uc.repository.Delete(id)
}
