package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetAllTeraluxUseCase handles the business logic for retrieving all teralux records
type GetAllTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewGetAllTeraluxUseCase creates a new instance of GetAllTeraluxUseCase
func NewGetAllTeraluxUseCase(repository *repositories.TeraluxRepository) *GetAllTeraluxUseCase {
	return &GetAllTeraluxUseCase{
		repository: repository,
	}
}

// Execute retrieves all teralux records
func (uc *GetAllTeraluxUseCase) Execute() (*dtos.TeraluxListResponseDTO, error) {
	// Get all teralux from repository
	teraluxList, err := uc.repository.GetAll()
	if err != nil {
		return nil, err
	}

	// Transform entities to DTOs
	teraluxDTOs := make([]dtos.TeraluxResponseDTO, len(teraluxList))
	for i, teralux := range teraluxList {
		teraluxDTOs[i] = dtos.TeraluxResponseDTO{
			ID:         teralux.ID,
			MacAddress: teralux.MacAddress,
			Name:       teralux.Name,
			CreatedAt:  teralux.CreatedAt,
			UpdatedAt:  teralux.UpdatedAt,
		}
	}

	return &dtos.TeraluxListResponseDTO{
		Teralux: teraluxDTOs,
		Total:   len(teraluxDTOs),
	}, nil
}
