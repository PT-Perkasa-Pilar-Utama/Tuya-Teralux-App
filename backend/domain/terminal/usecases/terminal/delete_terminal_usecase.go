package usecases

import (
	"errors"
	"regexp"
	"sensio/domain/terminal/repositories"
)

// DeleteTerminalUseCase handles deleting a terminal
type DeleteTerminalUseCase struct {
	repository repositories.ITerminalRepository
}

// NewDeleteTerminalUseCase creates a new instance of DeleteTerminalUseCase
func NewDeleteTerminalUseCase(repository repositories.ITerminalRepository) *DeleteTerminalUseCase {
	return &DeleteTerminalUseCase{
		repository: repository,
	}
}

func (uc *DeleteTerminalUseCase) DeleteTerminal(id string) error {
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return errors.New("Invalid ID format")
	}

	// Check existence
	if _, err := uc.repository.GetByID(id); err != nil {
		return errors.New("Terminal not found")
	}

	if err := uc.repository.Delete(id); err != nil {
		return err
	}

	return uc.repository.InvalidateCache(id)
}
