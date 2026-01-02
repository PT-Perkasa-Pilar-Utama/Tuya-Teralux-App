package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetAllDevicesUseCase handles retrieving all devices
type GetAllDevicesUseCase struct {
	repository DeviceRepository
}

// NewGetAllDevicesUseCase creates a new instance of GetAllDevicesUseCase
func NewGetAllDevicesUseCase(repository DeviceRepository) *GetAllDevicesUseCase {
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
			ID:        device.ID,
			TeraluxID: device.TeraluxID,
			Name:      device.Name,
			CreatedAt: device.CreatedAt,
			UpdatedAt: device.UpdatedAt,
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
