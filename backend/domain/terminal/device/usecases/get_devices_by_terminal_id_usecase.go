package usecases

import (
	"fmt"
	"sensio/domain/terminal/device/dtos"
	device_repositories "sensio/domain/terminal/device/repositories"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	"time"
)

// GetDevicesByTerminalIDUseCase handles retrieving devices linked to a terminal ID
type GetDevicesByTerminalIDUseCase struct {
	repository   device_repositories.IDeviceRepository
	terminalRepo terminal_repositories.ITerminalRepository
}

// NewGetDevicesByTerminalIDUseCase creates a new instance of GetDevicesByTerminalIDUseCase
func NewGetDevicesByTerminalIDUseCase(repository device_repositories.IDeviceRepository, terminalRepo terminal_repositories.ITerminalRepository) *GetDevicesByTerminalIDUseCase {
	return &GetDevicesByTerminalIDUseCase{
		repository:   repository,
		terminalRepo: terminalRepo,
	}
}

// Execute retrieves device records by terminal ID
func (uc *GetDevicesByTerminalIDUseCase) ListDevicesByTerminalID(terminalID string, page, limit int) (*dtos.DeviceListResponseDTO, error) {
	// 1. Check if Terminal ID exists
	_, err := uc.terminalRepo.GetByID(terminalID)
	if err != nil {
		return nil, fmt.Errorf("Terminal hub not found: %w", err)
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
	devices, total, err := uc.repository.GetByTerminalIDPaginated(terminalID, offset, limit)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	deviceDTOs := make([]dtos.DeviceResponseDTO, 0, len(devices))
	for _, item := range devices {
		deviceDTOs = append(deviceDTOs, dtos.DeviceResponseDTO{
			ID:                item.ID,
			TerminalID:        item.TerminalID,
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
			CreatedAt:         item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         item.UpdatedAt.Format(time.RFC3339),
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
