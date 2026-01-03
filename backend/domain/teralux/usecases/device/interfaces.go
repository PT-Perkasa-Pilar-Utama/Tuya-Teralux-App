package usecases

import (
	"teralux_app/domain/teralux/entities"
	tuya_dtos "teralux_app/domain/tuya/dtos"
)

// DeviceRepository defines the interface for device data access
type DeviceRepository interface {
	Create(device *entities.Device) error
	GetAll() ([]entities.Device, error)
	GetByTeraluxID(teraluxID string) ([]entities.Device, error)
	GetByID(id string) (*entities.Device, error)
	Update(device *entities.Device) error
	Delete(id string) error
}

type DeviceStatusRepository interface {
	UpsertDeviceStatuses(deviceID string, statuses []entities.DeviceStatus) error
	DeleteByDeviceID(deviceID string) error
}

// TuyaAuthUseCase defines the interface for Tuya authentication
type TuyaAuthUseCase interface {
	Authenticate() (*tuya_dtos.TuyaAuthResponseDTO, error)
}

// TuyaGetDeviceByIDUseCase defines the interface for fetching Tuya device details
type TuyaGetDeviceByIDUseCase interface {
	GetDeviceByID(accessToken, deviceID string) (*tuya_dtos.TuyaDeviceDTO, error)
}
