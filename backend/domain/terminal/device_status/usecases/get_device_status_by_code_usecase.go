package usecases

import (
	"fmt"
	"time"
	device_repositories "sensio/domain/terminal/device/repositories"
	"sensio/domain/terminal/device_status/dtos"
	device_status_repositories "sensio/domain/terminal/device_status/repositories"
)

// GetDeviceStatusByCodeUseCase handles retrieving a single device status by code
type GetDeviceStatusByCodeUseCase struct {
	repo    device_status_repositories.IDeviceStatusRepository
	devRepo device_repositories.IDeviceRepository
}

// NewGetDeviceStatusByCodeUseCase creates a new instance of GetDeviceStatusByCodeUseCase
func NewGetDeviceStatusByCodeUseCase(repo device_status_repositories.IDeviceStatusRepository, devRepo device_repositories.IDeviceRepository) *GetDeviceStatusByCodeUseCase {
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
			CreatedAt: status.CreatedAt.Format(time.RFC3339),
			UpdatedAt: status.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}
