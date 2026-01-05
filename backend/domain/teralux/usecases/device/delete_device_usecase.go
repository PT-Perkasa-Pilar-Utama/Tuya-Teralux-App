package usecases

import (
	"errors"
	"fmt"
	"regexp"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/repositories"
)

// DeleteDeviceUseCase handles deleting a device
type DeleteDeviceUseCase struct {
	repository       *repositories.DeviceRepository
	statusRepository *repositories.DeviceStatusRepository
	teraluxRepo      *repositories.TeraluxRepository
}

// NewDeleteDeviceUseCase creates a new instance of DeleteDeviceUseCase
func NewDeleteDeviceUseCase(
	repository *repositories.DeviceRepository,
	statusRepository *repositories.DeviceStatusRepository,
	teraluxRepo *repositories.TeraluxRepository,
) *DeleteDeviceUseCase {
	return &DeleteDeviceUseCase{
		repository:       repository,
		statusRepository: statusRepository,
		teraluxRepo:      teraluxRepo,
	}
}

// Execute deletes a device by ID and its associated statuses
func (uc *DeleteDeviceUseCase) Execute(id string) error {
	// Validation: Invalid ID Format
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return errors.New("Invalid ID format")
	}

	// Check existence and get device to find teralux_id
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

	// Invalidate teralux cache so next fetch gets fresh data without deleted device
	if err := uc.teraluxRepo.InvalidateCache(device.TeraluxID); err != nil {
		utils.LogWarn("DeleteDevice: Failed to invalidate teralux cache: %v", err)
	}

	return nil
}
