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
func (uc *GetDeviceStatusesByDeviceIDUseCase) Execute(deviceID string) (*dtos.DeviceStatusListResponseDTO, error) {
	_, err := uc.devRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("Device not found: %w", err)
	}

	statuses, err := uc.repo.GetByDeviceID(deviceID)
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

	return &dtos.DeviceStatusListResponseDTO{
		DeviceStatuses: dtosList,
		Total:          len(dtosList),
		Page:           1,
		PerPage:        len(dtosList),
	}, nil
}
