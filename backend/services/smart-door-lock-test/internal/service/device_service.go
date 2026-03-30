package service

import (
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"sensio/backend/services/smart-door-lock-test/internal/repository/tuya"
)

// DeviceService handles device-related business logic
type DeviceService struct {
	deviceRepo *tuya.DeviceRepository
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceRepo *tuya.DeviceRepository) *DeviceService {
	return &DeviceService{deviceRepo: deviceRepo}
}

// GetDevice retrieves a device with its specifications
func (s *DeviceService) GetDevice(deviceID string) (*domain.Device, *tuya.DeviceSpecifications, error) {
	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, nil, err
	}

	specs, err := s.deviceRepo.GetSpecifications(deviceID)
	if err != nil {
		// Specifications are optional, don't fail the whole operation
		specs = &tuya.DeviceSpecifications{}
	}

	return device, specs, nil
}

// GetLockState returns the current lock state
func (s *DeviceService) GetLockState(deviceID string) (domain.LockState, error) {
	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return domain.LockStateUnknown, err
	}

	return device.GetLockState(), nil
}

// IsOnline checks if the device is online
func (s *DeviceService) IsOnline(deviceID string) (bool, error) {
	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return false, err
	}

	return device.IsOnline(), nil
}
