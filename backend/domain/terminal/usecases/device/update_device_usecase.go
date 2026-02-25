package usecases

import (
	"errors"
	"strings"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/repositories"
)

// UpdateDeviceUseCase handles updating an existing device
type UpdateDeviceUseCase struct {
	repository  repositories.IDeviceRepository
	terminalRepo repositories.ITerminalRepository
}

// NewUpdateDeviceUseCase creates a new instance of UpdateDeviceUseCase
func NewUpdateDeviceUseCase(repository repositories.IDeviceRepository, terminalRepo repositories.ITerminalRepository) *UpdateDeviceUseCase {
	return &UpdateDeviceUseCase{
		repository:  repository,
		terminalRepo: terminalRepo,
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

	// Invalidate terminal cache
	return uc.terminalRepo.InvalidateCache(device.TerminalID)
}
