package usecases

import (
	"fmt"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetDevicesByTeraluxIDUseCase handles retrieving devices linked to a teralux ID
type GetDevicesByTeraluxIDUseCase struct {
	repository  *repositories.DeviceRepository
	teraluxRepo *repositories.TeraluxRepository
}

// NewGetDevicesByTeraluxIDUseCase creates a new instance of GetDevicesByTeraluxIDUseCase
func NewGetDevicesByTeraluxIDUseCase(repository *repositories.DeviceRepository, teraluxRepo *repositories.TeraluxRepository) *GetDevicesByTeraluxIDUseCase {
	return &GetDevicesByTeraluxIDUseCase{
		repository:  repository,
		teraluxRepo: teraluxRepo,
	}
}

// Execute retrieves device records by teralux ID
func (uc *GetDevicesByTeraluxIDUseCase) Execute(teraluxID string, page, limit int) (*dtos.DeviceListResponseDTO, error) {
	// 1. Check if Teralux ID exists
	_, err := uc.teraluxRepo.GetByID(teraluxID)
	if err != nil {
		return nil, fmt.Errorf("Teralux hub not found: %w", err)
	}

	// 2. Prepare Pagination
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// 3. Fetch devices
	devices, total, err := uc.repository.GetByTeraluxIDPaginated(teraluxID, offset, limit)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
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

	// If limit was 0 (meaning all), set per_page to total
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
