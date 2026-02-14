package usecases

import (
	"fmt"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetDeviceStatusByCodeUseCase handles retrieving a single device status by code
type GetDeviceStatusByCodeUseCase struct {
	repo    *repositories.DeviceStatusRepository
	devRepo *repositories.DeviceRepository
}

// NewGetDeviceStatusByCodeUseCase creates a new instance of GetDeviceStatusByCodeUseCase
func NewGetDeviceStatusByCodeUseCase(repo *repositories.DeviceStatusRepository, devRepo *repositories.DeviceRepository) *GetDeviceStatusByCodeUseCase {
	return &GetDeviceStatusByCodeUseCase{
		repo:    repo,
		devRepo: devRepo,
	}
}

// Execute retrieves a device status by device ID and code
func (uc *GetDeviceStatusByCodeUseCase) GetDeviceStatusByCode(deviceID string, code string) (*dtos.DeviceStatusSingleResponseDTO, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	_, err := uc.devRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("Device not found: %w", err)
	}

	status, err := uc.repo.GetByDeviceIDAndCode(deviceID, code)
	if err != nil {
		// assuming repo returns error if not found
		return nil, fmt.Errorf("Status code not found: %w", err)
	}
	if status.Code == "" {
		return nil, fmt.Errorf("Status code not found")
	}

	return &dtos.DeviceStatusSingleResponseDTO{
		DeviceStatus: dtos.DeviceStatusResponseDTO{
			DeviceID:  status.DeviceID,
			Code:      status.Code,
			Value:     status.Value,
			CreatedAt: status.CreatedAt,
			UpdatedAt: status.UpdatedAt,
		},
	}, nil
}
