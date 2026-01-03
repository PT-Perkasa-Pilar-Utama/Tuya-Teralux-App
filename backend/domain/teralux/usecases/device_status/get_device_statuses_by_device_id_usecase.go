package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetDeviceStatusesByDeviceIDUseCase handles retrieving all statuses for a specific device
type GetDeviceStatusesByDeviceIDUseCase struct {
	repository DeviceStatusRepository
}

// NewGetDeviceStatusesByDeviceIDUseCase creates a new instance of GetDeviceStatusesByDeviceIDUseCase
func NewGetDeviceStatusesByDeviceIDUseCase(repository DeviceStatusRepository) *GetDeviceStatusesByDeviceIDUseCase {
	return &GetDeviceStatusesByDeviceIDUseCase{
		repository: repository,
	}
}

// Execute retrieves all statuses for a device
func (uc *GetDeviceStatusesByDeviceIDUseCase) Execute(deviceID string) ([]dtos.DeviceStatusResponseDTO, error) {
	statuses, err := uc.repository.GetByDeviceID(deviceID)
	if err != nil {
		return nil, err
	}

	response := make([]dtos.DeviceStatusResponseDTO, len(statuses))
	for i, s := range statuses {
		response[i] = dtos.DeviceStatusResponseDTO{
			DeviceID:  s.DeviceID,
			Code:      s.Code,
			Value:     s.Value,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}

	return response, nil
}
