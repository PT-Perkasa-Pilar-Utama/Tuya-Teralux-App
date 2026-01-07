package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetAllDeviceStatusesUseCase handles retrieving all device statuses
type GetAllDeviceStatusesUseCase struct {
	repository *repositories.DeviceStatusRepository
}

// NewGetAllDeviceStatusesUseCase creates a new instance of GetAllDeviceStatusesUseCase
func NewGetAllDeviceStatusesUseCase(repository *repositories.DeviceStatusRepository) *GetAllDeviceStatusesUseCase {
	return &GetAllDeviceStatusesUseCase{
		repository: repository,
	}
}

// Execute retrieves all device statuses
func (uc *GetAllDeviceStatusesUseCase) Execute(page, limit int) (*dtos.DeviceStatusListResponseDTO, error) {
	// Prepare Pagination
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Fetch Data
	statuses, total, err := uc.repository.GetAllPaginated(offset, limit)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	var statusDTOs []dtos.DeviceStatusResponseDTO
	for _, status := range statuses {
		statusDTOs = append(statusDTOs, dtos.DeviceStatusResponseDTO{
			DeviceID:  status.DeviceID,
			Code:      status.Code,
			Value:     status.Value,
			CreatedAt: status.CreatedAt,
			UpdatedAt: status.UpdatedAt,
		})
	}

	// Ensure empty slice is not nil
	if statusDTOs == nil {
		statusDTOs = []dtos.DeviceStatusResponseDTO{}
	}

	// If limit was 0 (meaning all), set per_page to total
	if limit == 0 {
		limit = int(total)
	}

	return &dtos.DeviceStatusListResponseDTO{
		DeviceStatuses: statusDTOs,
		Total:          int(total),
		Page:           page,
		PerPage:        limit,
	}, nil
}
