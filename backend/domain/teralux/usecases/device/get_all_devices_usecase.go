package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetAllDevicesUseCase handles retrieving all devices
type GetAllDevicesUseCase struct {
	repository *repositories.DeviceRepository
}

// NewGetAllDevicesUseCase creates a new instance of GetAllDevicesUseCase
func NewGetAllDevicesUseCase(repository *repositories.DeviceRepository) *GetAllDevicesUseCase {
	return &GetAllDevicesUseCase{
		repository: repository,
	}
}

// Execute retrieves all devices
func (uc *GetAllDevicesUseCase) Execute() (*dtos.DeviceListResponseDTO, error) {
	devices, err := uc.repository.GetAll()
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	var deviceDTOs []dtos.DeviceResponseDTO
	for _, device := range devices {
		deviceDTOs = append(deviceDTOs, dtos.DeviceResponseDTO{
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
		})
	}

	// Ensure empty slice is not nil for JSON
	if deviceDTOs == nil {
		deviceDTOs = []dtos.DeviceResponseDTO{}
	}

	return &dtos.DeviceListResponseDTO{
		Devices: deviceDTOs,
		Total:   len(deviceDTOs),
	}, nil
}
