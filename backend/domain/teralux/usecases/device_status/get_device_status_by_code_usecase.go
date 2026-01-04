package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetDeviceStatusByCodeUseCase handles retrieving a single device status by code
type GetDeviceStatusByCodeUseCase struct {
	repository *repositories.DeviceStatusRepository
}

// NewGetDeviceStatusByCodeUseCase creates a new instance of GetDeviceStatusByCodeUseCase
func NewGetDeviceStatusByCodeUseCase(repository *repositories.DeviceStatusRepository) *GetDeviceStatusByCodeUseCase {
	return &GetDeviceStatusByCodeUseCase{
		repository: repository,
	}
}

// Execute retrieves a device status by device ID and code
func (uc *GetDeviceStatusByCodeUseCase) Execute(deviceID, code string) (*dtos.DeviceStatusResponseDTO, error) {
	status, err := uc.repository.GetByDeviceIDAndCode(deviceID, code)
	if err != nil {
		return nil, err
	}

	return &dtos.DeviceStatusResponseDTO{
		DeviceID:  status.DeviceID,
		Code:      status.Code,
		Value:     status.Value,
		CreatedAt: status.CreatedAt,
		UpdatedAt: status.UpdatedAt,
	}, nil
}
