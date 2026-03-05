package usecases

import (
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/repositories"
)

// GetAllTerminalUseCase handles retrieving all terminal records
type GetAllTerminalUseCase struct {
	repository repositories.ITerminalRepository
}

// NewGetAllTerminalUseCase creates a new instance of GetAllTerminalUseCase
func NewGetAllTerminalUseCase(repository repositories.ITerminalRepository) *GetAllTerminalUseCase {
	return &GetAllTerminalUseCase{
		repository: repository,
	}
}

// Execute retrieves all terminal records
func (uc *GetAllTerminalUseCase) ListTerminal(filter *dtos.TerminalFilterDTO) (*dtos.TerminalListResponseDTO, error) {
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
	terminalList, total, err := uc.repository.GetAllPaginated(offset, limit, roomID)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	terminalDTOs := make([]dtos.TerminalResponseDTO, 0, len(terminalList))
	for _, item := range terminalList {
		terminalDTOs = append(terminalDTOs, dtos.TerminalResponseDTO{
			ID:           item.ID,
			MacAddress:   item.MacAddress,
			Name:         item.Name,
			RoomID:       item.RoomID,
			DeviceTypeID: item.DeviceTypeID,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	if terminalDTOs == nil {
		terminalDTOs = []dtos.TerminalResponseDTO{}
	}

	// If limit was 0 (meaning all), set per_page to total
	if limit == 0 {
		limit = int(total)
	}

	return &dtos.TerminalListResponseDTO{
		Terminal: terminalDTOs,
		Total:    total,
		Page:     page,
		PerPage:  limit,
	}, nil
}
