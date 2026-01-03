package usecases

// DeleteDeviceUseCase handles deleting a device
type DeleteDeviceUseCase struct {
	repository       DeviceRepository
	statusRepository DeviceStatusRepository
}

// NewDeleteDeviceUseCase creates a new instance of DeleteDeviceUseCase
func NewDeleteDeviceUseCase(repository DeviceRepository, statusRepository DeviceStatusRepository) *DeleteDeviceUseCase {
	return &DeleteDeviceUseCase{
		repository:       repository,
		statusRepository: statusRepository,
	}
}

// Execute deletes a device by ID and its associated statuses
func (uc *DeleteDeviceUseCase) Execute(id string) error {
	// Delete associated statuses first
	if err := uc.statusRepository.DeleteByDeviceID(id); err != nil {
		return err
	}
	return uc.repository.Delete(id)
}
