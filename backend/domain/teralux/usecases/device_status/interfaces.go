package usecases

import (
	"teralux_app/domain/teralux/entities"
)

// DeviceStatusRepository defines the interface for device status data access
type DeviceStatusRepository interface {
	Create(status *entities.DeviceStatus) error
	GetAll() ([]entities.DeviceStatus, error)
	GetByDeviceID(deviceID string) ([]entities.DeviceStatus, error)
	GetByDeviceIDAndCode(deviceID, code string) (*entities.DeviceStatus, error)
	Upsert(status *entities.DeviceStatus) error
	DeleteByDeviceIDAndCode(deviceID, code string) error
	UpsertDeviceStatuses(deviceID string, statuses []entities.DeviceStatus) error
}
