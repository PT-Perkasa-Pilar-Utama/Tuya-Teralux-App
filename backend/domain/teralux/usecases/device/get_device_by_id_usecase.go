package usecases

import (
	"errors"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetDeviceByIDUseCase handles retrieving a single device
type GetDeviceByIDUseCase struct {
	repository repositories.IDeviceRepository
}

// NewGetDeviceByIDUseCase creates a new instance of GetDeviceByIDUseCase
func NewGetDeviceByIDUseCase(repository repositories.IDeviceRepository) *GetDeviceByIDUseCase {
	return &GetDeviceByIDUseCase{
		repository: repository,
	}
}

// Execute retrieves a device by ID
func (uc *GetDeviceByIDUseCase) GetDeviceByID(id string) (*dtos.DeviceSingleResponseDTO, error) {
	device, err := uc.repository.GetByID(id) // Kept 'device' as the variable name, assuming 'dev item' was a typo in the instruction
	if err != nil {
		return nil, errors.New("Device not found") // Changed error return
	}

	return &dtos.DeviceSingleResponseDTO{
		Device: dtos.DeviceResponseDTO{
			ID:                device.ID,
			TeraluxID:         device.TeraluxID,
			Name:              device.Name,
			RemoteID:          device.RemoteID,
			Category:          device.Category,
			RemoteCategory:    device.RemoteCategory,
			ProductName:       device.ProductName,
			RemoteProductName: device.RemoteProductName,
			Icon:              device.Icon,
			CustomName:        device.CustomName,
			Model:             device.Model,
			IP:                device.IP,
			LocalKey:          device.LocalKey,
			GatewayID:         device.GatewayID,
			CreateTime:        device.CreateTime,
			UpdateTime:        device.UpdateTime,
			Collections:       device.Collections,
			CreatedAt:         device.CreatedAt,
			UpdatedAt:         device.UpdatedAt,
		},
	}, nil
}
