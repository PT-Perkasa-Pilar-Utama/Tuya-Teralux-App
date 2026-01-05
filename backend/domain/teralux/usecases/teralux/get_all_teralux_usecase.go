package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
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
func (uc *GetAllTeraluxUseCase) Execute(filter *dtos.TeraluxFilterDTO) (*dtos.TeraluxListResponseDTO, error) {
	teraluxList, err := uc.repository.GetAll()
	if err != nil {
		return nil, err
	}

	// Filter in memory (Basic implementation to satisfy tests roughly)
	var filtered []entities.Teralux
	for _, item := range teraluxList {
		if filter != nil && filter.RoomID != nil {
			if item.RoomID != *filter.RoomID {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	// Pagination
	page := 1
	limit := 10
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 {
			limit = filter.Limit
		}
	}

	total := int64(len(filtered))
	start := (page - 1) * limit
	end := start + limit

	var paged []entities.Teralux
	if start < len(filtered) {
		if end > len(filtered) {
			end = len(filtered)
		}
		paged = filtered[start:end]
	}

	// Map to DTOs
	var teraluxDTOs []dtos.TeraluxResponseDTO
	for _, item := range paged {
		teraluxDTOs = append(teraluxDTOs, dtos.TeraluxResponseDTO{
			ID:         item.ID,
			MacAddress: item.MacAddress,
			Name:       item.Name,
			RoomID:     item.RoomID,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	if teraluxDTOs == nil {
		teraluxDTOs = []dtos.TeraluxResponseDTO{}
	}

	return &dtos.TeraluxListResponseDTO{
		Teralux: teraluxDTOs,
		Total:   total,
		Page:    page,
		PerPage: limit,
	}, nil
}
