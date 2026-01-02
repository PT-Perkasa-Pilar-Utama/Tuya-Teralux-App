package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// UpdateDeviceUseCase handles updating an existing device
type UpdateDeviceUseCase struct {
	repository DeviceRepository
}

// NewUpdateDeviceUseCase creates a new instance of UpdateDeviceUseCase
func NewUpdateDeviceUseCase(repository DeviceRepository) *UpdateDeviceUseCase {
	return &UpdateDeviceUseCase{
		repository: repository,
	}
}

// Execute updates a device
func (uc *UpdateDeviceUseCase) Execute(id string, req *dtos.UpdateDeviceRequestDTO) error {
	// First check if device exists
	device, err := uc.repository.GetByID(id)
	if err != nil {
		return err
	}

	// Update fields if present
	if req.Name != "" {
		device.Name = req.Name
	}

	// Save changes
	return uc.repository.Update(device)
}
