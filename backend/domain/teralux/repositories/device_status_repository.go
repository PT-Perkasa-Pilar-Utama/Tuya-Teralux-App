package repositories

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/teralux/entities"

	"gorm.io/gorm"
)

// DeviceStatusRepository handles database operations for DeviceStatus entities
type DeviceStatusRepository struct {
	db    *gorm.DB
	cache *persistence.BadgerService
}

// NewDeviceStatusRepository creates a new instance of DeviceStatusRepository
func NewDeviceStatusRepository(cache *persistence.BadgerService) *DeviceStatusRepository {
	return &DeviceStatusRepository{
		db:    infrastructure.DB,
		cache: cache,
	}
}

// Create inserts a new device status record into the database
func (r *DeviceStatusRepository) Create(status *entities.DeviceStatus) error {
	return r.db.Create(status).Error
}

// GetAll retrieves all active (non-deleted) device status records
func (r *DeviceStatusRepository) GetAll() ([]entities.DeviceStatus, error) {
	var statuses []entities.DeviceStatus
	err := r.db.Find(&statuses).Error
	return statuses, err
}

// GetByDeviceID retrieves all statuses belonging to a specific Device
func (r *DeviceStatusRepository) GetByDeviceID(deviceID string) ([]entities.DeviceStatus, error) {
	var statuses []entities.DeviceStatus
	err := r.db.Where("device_id = ?", deviceID).Find(&statuses).Error
	return statuses, err
}

// GetByDeviceIDAndCode retrieves a specific status by device ID and code
func (r *DeviceStatusRepository) GetByDeviceIDAndCode(deviceID, code string) (*entities.DeviceStatus, error) {
	var status entities.DeviceStatus
	err := r.db.Where("device_id = ? AND code = ?", deviceID, code).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// UpsertDeviceStatuses replaces all statuses for a device with new ones
func (r *DeviceStatusRepository) UpsertDeviceStatuses(deviceID string, statuses []entities.DeviceStatus) error {
	// Start transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete all existing statuses for this device
		if err := tx.Where("device_id = ?", deviceID).Delete(&entities.DeviceStatus{}).Error; err != nil {
			return err
		}

		// Insert new statuses
		if len(statuses) > 0 {
			if err := tx.Create(&statuses).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Upsert creates or updates a device status by composite key
func (r *DeviceStatusRepository) Upsert(status *entities.DeviceStatus) error {
	return r.db.Save(status).Error
}

// DeleteByDeviceIDAndCode deletes a device status by composite key
func (r *DeviceStatusRepository) DeleteByDeviceIDAndCode(deviceID, code string) error {
	return r.db.Where("device_id = ? AND code = ?", deviceID, code).Delete(&entities.DeviceStatus{}).Error
}
