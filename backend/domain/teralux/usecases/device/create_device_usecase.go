package usecases

import (
	"encoding/json"
	"fmt"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	tuya_usecases "teralux_app/domain/tuya/usecases"

	"github.com/google/uuid"
)

// CreateDeviceUseCase handles the business logic for creating a new device
type CreateDeviceUseCase struct {
	repository       *repositories.DeviceRepository
	statusRepository *repositories.DeviceStatusRepository
	tuyaAuthUC       *tuya_usecases.TuyaAuthUseCase
	tuyaGetDeviceUC  *tuya_usecases.TuyaGetDeviceByIDUseCase
}

// NewCreateDeviceUseCase creates a new instance of CreateDeviceUseCase
func NewCreateDeviceUseCase(
	repository *repositories.DeviceRepository,
	statusRepository *repositories.DeviceStatusRepository,
	tuyaAuthUC *tuya_usecases.TuyaAuthUseCase,
	tuyaGetDeviceUC *tuya_usecases.TuyaGetDeviceByIDUseCase,
) *CreateDeviceUseCase {
	return &CreateDeviceUseCase{
		repository:       repository,
		statusRepository: statusRepository,
		tuyaAuthUC:       tuyaAuthUC,
		tuyaGetDeviceUC:  tuyaGetDeviceUC,
	}
}

// Execute creates a new device record with automated status fetching from Tuya
func (uc *CreateDeviceUseCase) Execute(req *dtos.CreateDeviceRequestDTO) (*dtos.CreateDeviceResponseDTO, error) {
	// 1. Authenticate with Tuya to get token
	authResp, err := uc.tuyaAuthUC.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Tuya: %w", err)
	}

	// 2. Fetch all device details and status from Tuya
	tuyaDevice, err := uc.tuyaGetDeviceUC.GetDeviceByID(authResp.AccessToken, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch device from Tuya: %w", err)
	}

	// Check if device already exists by TeraluxID
	existingDevices, err := uc.repository.GetByTeraluxID(req.TeraluxID)
	if err != nil {
		return nil, err
	}

	var deviceID string
	var deviceEntity *entities.Device

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

	if len(existingDevices) > 0 {
		deviceEntity = &existingDevices[0]
		deviceID = deviceEntity.ID
		// Update existing record with latest Tuya info
		deviceEntity.Name = req.Name
		deviceEntity.RemoteID = tuyaDevice.ID
		deviceEntity.Category = tuyaDevice.Category
		deviceEntity.RemoteCategory = tuyaDevice.RemoteCategory
		deviceEntity.ProductName = tuyaDevice.ProductName
		deviceEntity.RemoteProductName = tuyaDevice.RemoteProductName
		deviceEntity.Icon = tuyaDevice.Icon
		deviceEntity.CustomName = tuyaDevice.CustomName
		deviceEntity.Model = tuyaDevice.Model
		deviceEntity.IP = tuyaDevice.IP
		deviceEntity.LocalKey = tuyaDevice.LocalKey
		deviceEntity.GatewayID = tuyaDevice.GatewayID
		deviceEntity.CreateTime = tuyaDevice.CreateTime
		deviceEntity.UpdateTime = tuyaDevice.UpdateTime
		deviceEntity.Collections = collectionsJSON

		if err := uc.repository.Update(deviceEntity); err != nil {
			return nil, err
		}
	} else {
		// Generate UUID for the new device
		deviceID = uuid.New().String()

		// Create entity
		deviceEntity = &entities.Device{
			ID:                deviceID,
			TeraluxID:         req.TeraluxID,
			Name:              req.Name,
			RemoteID:          tuyaDevice.ID,
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
			return nil, err
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
			return nil, err
		}
	}

	// Return response DTO with only ID
	return &dtos.CreateDeviceResponseDTO{
		ID: deviceID,
	}, nil
}
