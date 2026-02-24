package usecases

import (
	"errors"
	"regexp"
	"teralux_app/domain/teralux/repositories"
)

// DeleteTeraluxUseCase handles deleting a teralux
type DeleteTeraluxUseCase struct {
	repository repositories.ITeraluxRepository
}

// NewDeleteTeraluxUseCase creates a new instance of DeleteTeraluxUseCase
func NewDeleteTeraluxUseCase(repository repositories.ITeraluxRepository) *DeleteTeraluxUseCase {
	return &DeleteTeraluxUseCase{
		repository: repository,
	}
}

func (uc *DeleteTeraluxUseCase) DeleteTeralux(id string) error {
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return errors.New("Invalid ID format")
	}

	// Check existence
	if _, err := uc.repository.GetByID(id); err != nil {
		return errors.New("Teralux not found")
	}

	if err := uc.repository.Delete(id); err != nil {
		return err
	}

	return uc.repository.InvalidateCache(id)
}
