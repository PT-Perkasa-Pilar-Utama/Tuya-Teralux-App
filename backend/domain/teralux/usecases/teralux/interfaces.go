package usecases

import (
	"teralux_app/domain/teralux/entities"
)

// TeraluxRepository defines the interface for teralux data access
type TeraluxRepository interface {
	Create(teralux *entities.Teralux) error
	GetAll() ([]entities.Teralux, error)
	GetByID(id string) (*entities.Teralux, error)
	GetByMacAddress(macAddress string) (*entities.Teralux, error)
	Update(teralux *entities.Teralux) error
	Delete(id string) error
}

// DeviceRepository defines the interface for device data access
type DeviceRepository interface {
	GetByTeraluxID(teraluxID string) ([]entities.Device, error)
}
