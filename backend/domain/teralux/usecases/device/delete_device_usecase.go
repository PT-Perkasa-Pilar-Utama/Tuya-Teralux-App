package usecases

// DeleteDeviceUseCase handles deleting a device
type DeleteDeviceUseCase struct {
	repository DeviceRepository
}

// NewDeleteDeviceUseCase creates a new instance of DeleteDeviceUseCase
func NewDeleteDeviceUseCase(repository DeviceRepository) *DeleteDeviceUseCase {
	return &DeleteDeviceUseCase{
		repository: repository,
	}
}

// Execute deletes a device by ID
func (uc *DeleteDeviceUseCase) Execute(id string) error {
	return uc.repository.Delete(id)
}
