package usecases

// DeleteDeviceStatusUseCase handles deleting a device status
type DeleteDeviceStatusUseCase struct {
	repository DeviceStatusRepository
}

// NewDeleteDeviceStatusUseCase creates a new instance of DeleteDeviceStatusUseCase
func NewDeleteDeviceStatusUseCase(repository DeviceStatusRepository) *DeleteDeviceStatusUseCase {
	return &DeleteDeviceStatusUseCase{
		repository: repository,
	}
}

// Execute deletes a device status by ID
func (uc *DeleteDeviceStatusUseCase) Execute(id string) error {
	return uc.repository.Delete(id)
}
