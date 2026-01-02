package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"

	"github.com/google/uuid"
)

// CreateTeraluxUseCase handles the business logic for creating a new teralux
type CreateTeraluxUseCase struct {
	repository TeraluxRepository
}

// NewCreateTeraluxUseCase creates a new instance of CreateTeraluxUseCase
func NewCreateTeraluxUseCase(repository TeraluxRepository) *CreateTeraluxUseCase {
	return &CreateTeraluxUseCase{
		repository: repository,
	}
}

// Execute creates a new teralux record
func (uc *CreateTeraluxUseCase) Execute(req *dtos.CreateTeraluxRequestDTO) (*dtos.CreateTeraluxResponseDTO, error) {
	// Generate UUID for the new teralux
	id := uuid.New().String()

	// Create entity
	teralux := &entities.Teralux{
		ID:         id,
		MacAddress: req.MacAddress,
		Name:       req.Name,
	}

	// Save to database
	if err := uc.repository.Create(teralux); err != nil {
		return nil, err
	}

	// Return response DTO with only ID
	return &dtos.CreateTeraluxResponseDTO{
		ID: teralux.ID,
	}, nil
}
