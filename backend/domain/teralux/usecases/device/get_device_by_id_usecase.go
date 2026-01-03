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
	}, nil
}
