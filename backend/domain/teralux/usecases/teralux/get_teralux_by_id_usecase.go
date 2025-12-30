package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetTeraluxByIDUseCase handles the business logic for retrieving a single teralux by ID
type GetTeraluxByIDUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewGetTeraluxByIDUseCase creates a new instance of GetTeraluxByIDUseCase
func NewGetTeraluxByIDUseCase(repository *repositories.TeraluxRepository) *GetTeraluxByIDUseCase {
	return &GetTeraluxByIDUseCase{
		repository: repository,
	}
}

// Execute retrieves a single teralux by ID
func (uc *GetTeraluxByIDUseCase) Execute(id string) (*dtos.TeraluxResponseDTO, error) {
	// Get teralux from repository
	teralux, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Transform entity to DTO
	return &dtos.TeraluxResponseDTO{
		ID:         teralux.ID,
		MacAddress: teralux.MacAddress,
		Name:       teralux.Name,
		CreatedAt:  teralux.CreatedAt,
		UpdatedAt:  teralux.UpdatedAt,
	}, nil
}
