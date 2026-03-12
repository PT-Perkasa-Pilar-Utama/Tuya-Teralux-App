package usecases

import (
	"errors"
	"fmt"
	"regexp"
	"sensio/domain/common/utils"
	device_repositories "sensio/domain/terminal/device/repositories"
	device_status_repositories "sensio/domain/terminal/device_status/repositories"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
)

// DeleteDeviceUseCase handles deleting a device
type DeleteDeviceUseCase struct {
	repository       device_repositories.IDeviceRepository
	statusRepository device_status_repositories.IDeviceStatusRepository
	terminalRepo     terminal_repositories.ITerminalRepository
}

// NewDeleteDeviceUseCase creates a new instance of DeleteDeviceUseCase
func NewDeleteDeviceUseCase(
	repository device_repositories.IDeviceRepository,
	statusRepository device_status_repositories.IDeviceStatusRepository,
	terminalRepo terminal_repositories.ITerminalRepository,
) *DeleteDeviceUseCase {
	return &DeleteDeviceUseCase{
		repository:       repository,
		statusRepository: statusRepository,
		terminalRepo:     terminalRepo,
	}
}

// Execute deletes a device by ID and its associated statuses
func (uc *DeleteDeviceUseCase) DeleteDevice(id string) error {
	// Validation: Invalid ID Format
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return errors.New("Invalid ID format")
	}

	// Check existence and get device to find terminal_id
	device, err := uc.repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("Device not found: %w", err)
	}

	// Delete associated statuses first
	if err := uc.statusRepository.DeleteByDeviceID(id); err != nil {
		return err
	}

	// Delete device
	if err := uc.repository.Delete(id); err != nil {
		return err
	}

	// Invalidate terminal cache so next fetch gets fresh data without deleted device
	if err := uc.terminalRepo.InvalidateCache(device.TerminalID); err != nil {
		utils.LogWarn("DeleteDevice: Failed to invalidate terminal cache: %v", err)
	}

	return nil
}
