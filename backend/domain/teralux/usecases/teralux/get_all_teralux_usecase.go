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
func (uc *GetAllTeraluxUseCase) Execute(filter *dtos.TeraluxFilterDTO) (*dtos.TeraluxListResponseDTO, error) {
	// Prepare Pagination & Filter
	page := 1
	limit := 0
	var roomID *string

	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 {
			limit = filter.Limit
		} else if filter.PerPage > 0 {
			limit = filter.PerPage
		}
		roomID = filter.RoomID
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Fetch Data
	teraluxList, total, err := uc.repository.GetAllPaginated(offset, limit, roomID)
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
			RoomID:     item.RoomID,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	if teraluxDTOs == nil {
		teraluxDTOs = []dtos.TeraluxResponseDTO{}
	}

	// If limit was 0 (meaning all), set per_page to total
	if limit == 0 {
		limit = int(total)
	}

	return &dtos.TeraluxListResponseDTO{
		Teralux: teraluxDTOs,
		Total:   total,
		Page:    page,
		PerPage: limit,
	}, nil
}
