package usecases

import (
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// UpdateDeviceUseCase handles updating an existing device
type UpdateDeviceUseCase struct {
	repository *repositories.DeviceRepository
}

// NewUpdateDeviceUseCase creates a new instance of UpdateDeviceUseCase
func NewUpdateDeviceUseCase(repository *repositories.DeviceRepository) *UpdateDeviceUseCase {
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

	// Update fields
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
				{Field: "name", Message: "name cannot be empty"},
			})
		}
		device.Name = *req.Name
	}

	// Save changes
	return uc.repository.Update(device)
}
