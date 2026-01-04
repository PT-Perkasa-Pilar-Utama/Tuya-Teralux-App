package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
)

// GetTeraluxByIDUseCase handles retrieving a single teralux
type GetTeraluxByIDUseCase struct {
	repository *repositories.TeraluxRepository
	devRepo    *repositories.DeviceRepository
}

// NewGetTeraluxByIDUseCase creates a new instance of GetTeraluxByIDUseCase
func NewGetTeraluxByIDUseCase(repository *repositories.TeraluxRepository, devRepo *repositories.DeviceRepository) *GetTeraluxByIDUseCase {
	return &GetTeraluxByIDUseCase{
		repository: repository,
		devRepo:    devRepo,
	}
}

// Execute retrieves a teralux by ID with its associated devices
func (uc *GetTeraluxByIDUseCase) Execute(id string) (*dtos.TeraluxResponseDTO, error) {
	item, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Fetch devices associated with this teralux
	devices, err := uc.devRepo.GetByTeraluxID(id)
	if err != nil {
		// If error fetching devices, return teralux with empty devices array
		devices = []entities.Device{}
	}

	// Convert devices to DTOs
	deviceDTOs := make([]dtos.DeviceResponseDTO, len(devices))
	for i, device := range devices {
		deviceDTOs[i] = dtos.DeviceResponseDTO{
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
		}
	}

	return &dtos.TeraluxResponseDTO{
		ID:         item.ID,
		MacAddress: item.MacAddress,
		RoomID:     item.RoomID,
		Name:       item.Name,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
		Devices:    deviceDTOs,
	}, nil
}
