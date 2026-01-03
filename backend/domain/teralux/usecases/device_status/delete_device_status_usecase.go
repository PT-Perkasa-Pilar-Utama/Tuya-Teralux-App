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

// Execute deletes a device status by device ID and code
func (uc *DeleteDeviceStatusUseCase) Execute(deviceID, code string) error {
	return uc.repository.DeleteByDeviceIDAndCode(deviceID, code)
}
