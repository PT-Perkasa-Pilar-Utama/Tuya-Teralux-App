package repositories

import (
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/teralux/entities"

	"gorm.io/gorm"
)

// DeviceStatusRepository handles database operations for DeviceStatus entities
type DeviceStatusRepository struct {
	db    *gorm.DB
	cache *infrastructure.BadgerService
}

// NewDeviceStatusRepository creates a new instance of DeviceStatusRepository
func NewDeviceStatusRepository(cache *infrastructure.BadgerService) *DeviceStatusRepository {
	return &DeviceStatusRepository{
		db:    infrastructure.DB,
		cache: cache,
	}
}

// Create inserts a new device status record into the database
func (r *DeviceStatusRepository) Create(status *entities.DeviceStatus) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Create(status).Error
}

// GetAll retrieves all active (non-deleted) device status records
func (r *DeviceStatusRepository) GetAll() ([]entities.DeviceStatus, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var statuses []entities.DeviceStatus
	err := r.db.Find(&statuses).Error
	return statuses, err
}

// GetAllPaginated retrieves all active device status records with pagination
func (r *DeviceStatusRepository) GetAllPaginated(offset, limit int) ([]entities.DeviceStatus, int64, error) {
	if r.db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}
	var statuses []entities.DeviceStatus
	var total int64

	// Count total
	if err := r.db.Model(&entities.DeviceStatus{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated
	query := r.db
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	err := query.Find(&statuses).Error
	return statuses, total, err
}

// GetByDeviceID retrieves all statuses belonging to a specific Device
func (r *DeviceStatusRepository) GetByDeviceID(deviceID string) ([]entities.DeviceStatus, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var statuses []entities.DeviceStatus
	err := r.db.Where("device_id = ?", deviceID).Find(&statuses).Error
	return statuses, err
}

// GetByDeviceIDPaginated retrieves statuses by Device ID with pagination
func (r *DeviceStatusRepository) GetByDeviceIDPaginated(deviceID string, offset, limit int) ([]entities.DeviceStatus, int64, error) {
	if r.db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}
	var statuses []entities.DeviceStatus
	var total int64

	// Count total
	if err := r.db.Model(&entities.DeviceStatus{}).Where("device_id = ?", deviceID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated
	query := r.db.Where("device_id = ?", deviceID)
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	err := query.Find(&statuses).Error
	return statuses, total, err
}

// GetByDeviceIDAndCode retrieves a specific status by device ID and code
func (r *DeviceStatusRepository) GetByDeviceIDAndCode(deviceID, code string) (*entities.DeviceStatus, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var status entities.DeviceStatus
	err := r.db.Where("device_id = ? AND code = ?", deviceID, code).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// UpsertDeviceStatuses replaces all statuses for a device with new ones
func (r *DeviceStatusRepository) UpsertDeviceStatuses(deviceID string, statuses []entities.DeviceStatus) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	// Start transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete all existing statuses for this device (Hard Delete to avoid Unique Composite Key issues)
		if err := tx.Unscoped().Where("device_id = ?", deviceID).Delete(&entities.DeviceStatus{}).Error; err != nil {
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
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Save(status).Error
}

// DeleteByDeviceIDAndCode deletes a device status by composite key
func (r *DeviceStatusRepository) DeleteByDeviceIDAndCode(deviceID, code string) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Where("device_id = ? AND code = ?", deviceID, code).Delete(&entities.DeviceStatus{}).Error
}

// DeleteByDeviceID deletes all statuses for a specific device
func (r *DeviceStatusRepository) DeleteByDeviceID(deviceID string) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Where("device_id = ?", deviceID).Delete(&entities.DeviceStatus{}).Error
}
