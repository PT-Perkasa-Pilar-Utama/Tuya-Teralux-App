package usecases

import (
	"errors"
	"regexp"
	"teralux_app/domain/teralux/repositories"
)

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

func (uc *DeleteTeraluxUseCase) Execute(id string) error {
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return errors.New("Invalid ID format")
	}
	return uc.repository.Delete(id)
}
