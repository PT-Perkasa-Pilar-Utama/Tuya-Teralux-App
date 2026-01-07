package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
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

// Execute retrieves all device records
func (uc *GetAllDevicesUseCase) Execute(filter *dtos.DeviceFilterDTO) (*dtos.DeviceListResponseDTO, error) {
	// 1. Prepare Pagination
	page := 1
	limit := 0

	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 {
			limit = filter.Limit
		} else if filter.PerPage > 0 {
			limit = filter.PerPage
		}
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// 2. Fetch Data
	var devices []entities.Device
	var total int64
	var err error

	if filter != nil && filter.TeraluxID != nil && *filter.TeraluxID != "" {
		devices, total, err = uc.repository.GetByTeraluxIDPaginated(*filter.TeraluxID, offset, limit)
	} else {
		devices, total, err = uc.repository.GetAllPaginated(offset, limit)
	}

	if err != nil {
		return nil, err
	}

	// 3. Map to DTOs
	var deviceDTOs []dtos.DeviceResponseDTO
	for _, item := range devices {
		deviceDTOs = append(deviceDTOs, dtos.DeviceResponseDTO{
			ID:                item.ID,
			TeraluxID:         item.TeraluxID,
			Name:              item.Name,
			RemoteID:          item.RemoteID,
			Category:          item.Category,
			RemoteCategory:    item.RemoteCategory,
			ProductName:       item.ProductName,
			RemoteProductName: item.RemoteProductName,
			Icon:              item.Icon,
			CustomName:        item.CustomName,
			Model:             item.Model,
			IP:                item.IP,
			LocalKey:          item.LocalKey,
			GatewayID:         item.GatewayID,
			CreateTime:        item.CreateTime,
			UpdateTime:        item.UpdateTime,
			Collections:       item.Collections,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt,
		})
	}

	// Ensure empty slice is not nil for JSON
	if deviceDTOs == nil {
		deviceDTOs = []dtos.DeviceResponseDTO{}
	}

	// If limit was 0 (meaning all), set per_page to total for consistency
	if limit == 0 {
		limit = int(total)
	}

	return &dtos.DeviceListResponseDTO{
		Devices: deviceDTOs,
		Total:   int(total),
		Page:    page,
		PerPage: limit,
	}, nil
}
