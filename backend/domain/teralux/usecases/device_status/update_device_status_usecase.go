package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// UpdateDeviceStatusUseCase handles updating an existing device status
type UpdateDeviceStatusUseCase struct {
	repository DeviceStatusRepository
}

// NewUpdateDeviceStatusUseCase creates a new instance of UpdateDeviceStatusUseCase
func NewUpdateDeviceStatusUseCase(repository DeviceStatusRepository) *UpdateDeviceStatusUseCase {
	return &UpdateDeviceStatusUseCase{
		repository: repository,
	}
}

// Execute updates a device status
func (uc *UpdateDeviceStatusUseCase) Execute(deviceID, code string, req *dtos.UpdateDeviceStatusRequestDTO) error {
	// First check if status exists
	status, err := uc.repository.GetByDeviceIDAndCode(deviceID, code)
	if err != nil {
		return err
	}

	// Update fields if present
	if req.Value != "" {
		status.Value = req.Value
	}

	// Save changes
	return uc.repository.Upsert(status)
}
