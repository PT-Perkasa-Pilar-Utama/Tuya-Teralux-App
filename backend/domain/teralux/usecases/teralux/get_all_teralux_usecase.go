package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetAllTeraluxUseCase handles retrieving all teralux records
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
	teraluxList, err := uc.repository.GetAll()
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	var teraluxDTOs []dtos.TeraluxResponseDTO
	for _, item := range teraluxList {
		teraluxDTOs = append(teraluxDTOs, dtos.TeraluxResponseDTO{
			ID:         item.ID,
			MacAddress: item.MacAddress,
			Name:       item.Name,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	if teraluxDTOs == nil {
		teraluxDTOs = []dtos.TeraluxResponseDTO{}
	}

	return &dtos.TeraluxListResponseDTO{
		Teralux: teraluxDTOs,
		Total:   len(teraluxDTOs),
	}, nil
}
