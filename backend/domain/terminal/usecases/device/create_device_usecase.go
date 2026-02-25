package usecases

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/entities"
	"sensio/domain/terminal/repositories"
	tuya_dtos "sensio/domain/tuya/dtos"

	"gorm.io/gorm"
)

// TuyaAuthUseCaseInterface defines the interface for Tuya authentication
type TuyaAuthUseCaseInterface interface {
	Authenticate() (*tuya_dtos.TuyaAuthResponseDTO, error)
}

// TuyaGetDeviceByIDUseCaseInterface defines the interface for getting a device by ID from Tuya
type TuyaGetDeviceByIDUseCaseInterface interface {
	GetDeviceByID(accessToken, deviceID string) (*tuya_dtos.TuyaDeviceDTO, error)
}

// CreateDeviceUseCase handles the business logic for creating a new device
type CreateDeviceUseCase struct {
	repository       repositories.IDeviceRepository
	statusRepository repositories.IDeviceStatusRepository
	terminalRepo      repositories.ITerminalRepository
	tuyaAuthUC       TuyaAuthUseCaseInterface
	tuyaGetDeviceUC  TuyaGetDeviceByIDUseCaseInterface
}

// NewCreateDeviceUseCase creates a new instance of CreateDeviceUseCase
func NewCreateDeviceUseCase(
	repository repositories.IDeviceRepository,
	statusRepository repositories.IDeviceStatusRepository,
	tuyaAuthUC TuyaAuthUseCaseInterface,
	tuyaGetDeviceUC TuyaGetDeviceByIDUseCaseInterface,
	terminalRepo repositories.ITerminalRepository,
) *CreateDeviceUseCase {
	return &CreateDeviceUseCase{
		repository:       repository,
		statusRepository: statusRepository,
		terminalRepo:      terminalRepo,
		tuyaAuthUC:       tuyaAuthUC,
		tuyaGetDeviceUC:  tuyaGetDeviceUC,
	}
}

// Execute creates a new device record with automated status fetching from Tuya
func (uc *CreateDeviceUseCase) CreateDevice(req *dtos.CreateDeviceRequestDTO) (*dtos.CreateDeviceResponseDTO, bool, error) {
	// Validation
	var details []utils.ValidationErrorDetail
	if req.Name == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "name", Message: "name is required"})
	}
	if req.TerminalID == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "terminal_id", Message: "terminal_id is required"})
	}
	if len(details) > 0 {
		return nil, false, utils.NewValidationError("Validation Error", details)
	}

	// Constraint: Invalid Terminal ID
	_, err := uc.terminalRepo.GetByID(req.TerminalID)
	if err != nil {
		return nil, false, fmt.Errorf("Invalid terminal_id: Terminal hub does not exist")
	}

	// 1. Authenticate with Tuya to get token
	authResp, err := uc.tuyaAuthUC.Authenticate()
	if err != nil {
		return nil, false, fmt.Errorf("failed to authenticate with Tuya: %w", err)
	}

	tuyaDevice, err := uc.tuyaGetDeviceUC.GetDeviceByID(authResp.AccessToken, req.ID)

	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch device from Tuya: %w", err)
	}

	// Serialize collections if any
	collectionsJSON := "[]"
	if len(tuyaDevice.Collections) > 0 {
		ids := make([]string, len(tuyaDevice.Collections))
		for i, c := range tuyaDevice.Collections {
			ids[i] = c.ID
		}
		data, _ := json.Marshal(ids)
		collectionsJSON = string(data)
	}

	var deviceID string
	var deviceEntity *entities.Device

	// Check if device with this ID already exists (including soft-deleted)
	existingDevice, err := uc.repository.GetByIDUnscoped(req.ID)
	if err == nil && existingDevice != nil {
		// Device exists (Active or Soft-Deleted)
		utils.LogDebug("CreateDevice: Device %s found (DeletedAt: %v). Updating...", req.ID, existingDevice.DeletedAt)

		// Prepare update
		existingDevice.TerminalID = req.TerminalID
		existingDevice.Name = req.Name
		existingDevice.RemoteID = tuyaDevice.RemoteID
		existingDevice.Category = tuyaDevice.Category
		existingDevice.RemoteCategory = tuyaDevice.RemoteCategory
		existingDevice.ProductName = tuyaDevice.ProductName
		existingDevice.RemoteProductName = tuyaDevice.RemoteProductName
		existingDevice.Icon = tuyaDevice.Icon
		existingDevice.CustomName = tuyaDevice.CustomName
		existingDevice.Model = tuyaDevice.Model
		existingDevice.IP = tuyaDevice.IP
		existingDevice.LocalKey = tuyaDevice.LocalKey
		existingDevice.GatewayID = tuyaDevice.GatewayID
		existingDevice.CreateTime = tuyaDevice.CreateTime
		existingDevice.UpdateTime = tuyaDevice.UpdateTime
		existingDevice.Collections = collectionsJSON
		existingDevice.DeletedAt = gorm.DeletedAt{} // Restore if deleted

		if err := uc.repository.Update(existingDevice); err != nil {
			return nil, false, err
		}
		deviceID = existingDevice.ID
	} else {
		// Create new device
		deviceID = req.ID

		deviceEntity = &entities.Device{
			ID:                deviceID,
			TerminalID:         req.TerminalID,
			Name:              req.Name,
			RemoteID:          tuyaDevice.RemoteID,
			Category:          tuyaDevice.Category,
			RemoteCategory:    tuyaDevice.RemoteCategory,
			ProductName:       tuyaDevice.ProductName,
			RemoteProductName: tuyaDevice.RemoteProductName,
			Icon:              tuyaDevice.Icon,
			CustomName:        tuyaDevice.CustomName,
			Model:             tuyaDevice.Model,
			IP:                tuyaDevice.IP,
			LocalKey:          tuyaDevice.LocalKey,
			GatewayID:         tuyaDevice.GatewayID,
			CreateTime:        tuyaDevice.CreateTime,
			UpdateTime:        tuyaDevice.UpdateTime,
			Collections:       collectionsJSON,
		}

		// Save to database
		if err := uc.repository.Create(deviceEntity); err != nil {
			return nil, false, err
		}
	}

	// 3. Automatically map and Upsert statuses returned from Tuya
	if len(tuyaDevice.Status) > 0 {
		statusEntities := make([]entities.DeviceStatus, len(tuyaDevice.Status))
		for i, s := range tuyaDevice.Status {
			// Tuya status value can be anything, we store it as string
			valStr := fmt.Sprintf("%v", s.Value)
			statusEntities[i] = entities.DeviceStatus{
				DeviceID: deviceID,
				Code:     s.Code,
				Value:    valStr,
			}
		}

		if err := uc.statusRepository.UpsertDeviceStatuses(deviceID, statusEntities); err != nil {
			return nil, false, err
		}
	}

	// Invalidate terminal cache so next fetch gets fresh data with new device
	if err := uc.terminalRepo.InvalidateCache(req.TerminalID); err != nil {
		utils.LogWarn("CreateDevice: Failed to invalidate terminal cache: %v", err)
	}

	// Return response DTO with only ID and isNew flag (always true now)
	return &dtos.CreateDeviceResponseDTO{
		DeviceID: deviceID,
	}, true, nil
}
