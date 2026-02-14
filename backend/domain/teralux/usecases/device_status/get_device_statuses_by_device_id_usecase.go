package usecases

import (
	"fmt"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetDeviceStatusesByDeviceIDUseCase handles retrieving all statuses for a specific device
type GetDeviceStatusesByDeviceIDUseCase struct {
	repo    *repositories.DeviceStatusRepository
	devRepo *repositories.DeviceRepository
}

// NewGetDeviceStatusesByDeviceIDUseCase creates a new instance of GetDeviceStatusesByDeviceIDUseCase
func NewGetDeviceStatusesByDeviceIDUseCase(repo *repositories.DeviceStatusRepository, devRepo *repositories.DeviceRepository) *GetDeviceStatusesByDeviceIDUseCase {
	return &GetDeviceStatusesByDeviceIDUseCase{
		repo:    repo,
		devRepo: devRepo,
	}
}

// Execute retrieves all statuses for a device
func (uc *GetDeviceStatusesByDeviceIDUseCase) ListDeviceStatusesByDeviceID(deviceID string, page, limit int) (*dtos.DeviceStatusListResponseDTO, error) {
	_, err := uc.devRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("Device not found: %w", err)
	}

	// Prepare Pagination
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Fetch Data
	statuses, total, err := uc.repo.GetByDeviceIDPaginated(deviceID, offset, limit)
	if err != nil {
		return nil, err
	}

	var dtosList []dtos.DeviceStatusResponseDTO
	for _, s := range statuses {
		dtosList = append(dtosList, dtos.DeviceStatusResponseDTO{
			DeviceID:  s.DeviceID,
			Code:      s.Code,
			Value:     s.Value,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}

	if dtosList == nil {
		dtosList = []dtos.DeviceStatusResponseDTO{}
	}

	// If limit was 0 (meaning all), set per_page to total
	if limit == 0 {
		limit = int(total)
	}

	return &dtos.DeviceStatusListResponseDTO{
		DeviceStatuses: dtosList,
		Total:          int(total),
		Page:           page,
		PerPage:        limit,
	}, nil
}
