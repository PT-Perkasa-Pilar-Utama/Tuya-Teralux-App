package usecases

import (
	"teralux_app/domain/teralux/entities"
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
