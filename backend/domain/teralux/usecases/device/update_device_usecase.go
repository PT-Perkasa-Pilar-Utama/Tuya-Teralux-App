package usecases

import (
	"errors"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// UpdateDeviceUseCase handles updating an existing device
type UpdateDeviceUseCase struct {
	repository  repositories.IDeviceRepository
	teraluxRepo repositories.ITeraluxRepository
}

// NewUpdateDeviceUseCase creates a new instance of UpdateDeviceUseCase
func NewUpdateDeviceUseCase(repository repositories.IDeviceRepository, teraluxRepo repositories.ITeraluxRepository) *UpdateDeviceUseCase {
	return &UpdateDeviceUseCase{
		repository:  repository,
		teraluxRepo: teraluxRepo,
	}
}

// Execute updates a device
func (uc *UpdateDeviceUseCase) UpdateDevice(id string, req *dtos.UpdateDeviceRequestDTO) error {
	// First check if device exists
	device, err := uc.repository.GetByID(id)
	if err != nil {
		return errors.New("Device not found")
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
	if err := uc.repository.Update(device); err != nil {
		return err
	}

	// Invalidate teralux cache
	return uc.teraluxRepo.InvalidateCache(device.TeraluxID)
}
