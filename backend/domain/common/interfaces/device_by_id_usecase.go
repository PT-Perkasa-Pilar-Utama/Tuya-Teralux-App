package interfaces

import (
	"sensio/domain/tuya/dtos"
)

// DeviceByIDUseCase defines the interface for fetching device by ID
type DeviceByIDUseCase interface {
	GetDeviceByID(accessToken, deviceID, remoteID string) (*dtos.TuyaDeviceDTO, error)
}