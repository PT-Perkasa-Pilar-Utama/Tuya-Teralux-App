package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetDeviceStatusByIDUseCase handles retrieving a single device status
type GetDeviceStatusByIDUseCase struct {
	repository DeviceStatusRepository
}

// NewGetDeviceStatusByIDUseCase creates a new instance of GetDeviceStatusByIDUseCase
func NewGetDeviceStatusByIDUseCase(repository DeviceStatusRepository) *GetDeviceStatusByIDUseCase {
	return &GetDeviceStatusByIDUseCase{
		repository: repository,
	}
}

// Execute retrieves a device status by ID
func (uc *GetDeviceStatusByIDUseCase) Execute(id string) (*dtos.DeviceStatusResponseDTO, error) {
	status, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &dtos.DeviceStatusResponseDTO{
		ID:        status.ID,
		DeviceID:  status.DeviceID,
		Name:      status.Name,
		Code:      status.Code,
		Value:     status.Value,
		CreatedAt: status.CreatedAt,
		UpdatedAt: status.UpdatedAt,
	}, nil
}
