package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetDeviceByIDUseCase handles retrieving a single device
type GetDeviceByIDUseCase struct {
	repository *repositories.DeviceRepository
}

// NewGetDeviceByIDUseCase creates a new instance of GetDeviceByIDUseCase
func NewGetDeviceByIDUseCase(repository *repositories.DeviceRepository) *GetDeviceByIDUseCase {
	return &GetDeviceByIDUseCase{
		repository: repository,
	}
}

// Execute retrieves a device by ID
func (uc *GetDeviceByIDUseCase) Execute(id string) (*dtos.DeviceSingleResponseDTO, error) {
	device, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
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
