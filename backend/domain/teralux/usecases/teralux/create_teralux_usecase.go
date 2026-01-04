package usecases

import (
	"errors"
	"strings"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"

	"github.com/google/uuid"
)

// CreateTeraluxUseCase handles the business logic for creating a new teralux
type CreateTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewCreateTeraluxUseCase creates a new instance of CreateTeraluxUseCase
func NewCreateTeraluxUseCase(repository *repositories.TeraluxRepository) *CreateTeraluxUseCase {
	return &CreateTeraluxUseCase{
		repository: repository,
	}
}

// Execute creates a new teralux record
func (uc *CreateTeraluxUseCase) Execute(req *dtos.CreateTeraluxRequestDTO) (*dtos.CreateTeraluxResponseDTO, error) {
	// Validation
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(req.MacAddress) == "" {
		return nil, errors.New("mac_address is required")
	}
	if strings.TrimSpace(req.RoomID) == "" {
		return nil, errors.New("room_id is required")
	}

	// Generate UUID for the new teralux
	id := uuid.New().String()

	// Create entity
	teralux := &entities.Teralux{
		ID:         id,
		MacAddress: req.MacAddress,
		RoomID:     req.RoomID,
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
