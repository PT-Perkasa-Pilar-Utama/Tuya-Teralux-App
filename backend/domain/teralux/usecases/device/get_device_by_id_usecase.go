package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetDeviceByIDUseCase handles retrieving a single device
type GetDeviceByIDUseCase struct {
	repository DeviceRepository
}

// NewGetDeviceByIDUseCase creates a new instance of GetDeviceByIDUseCase
func NewGetDeviceByIDUseCase(repository DeviceRepository) *GetDeviceByIDUseCase {
	return &GetDeviceByIDUseCase{
		repository: repository,
	}
}

// Execute retrieves a device by ID
func (uc *GetDeviceByIDUseCase) Execute(id string) (*dtos.DeviceResponseDTO, error) {
	device, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &dtos.DeviceResponseDTO{
		ID:        device.ID,
		TeraluxID: device.TeraluxID,
		Name:      device.Name,
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
	}, nil
}
