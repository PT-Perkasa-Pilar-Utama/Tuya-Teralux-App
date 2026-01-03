package repositories

import (
	"encoding/json"
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/common/utils"
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

// GetByID retrieves a single device status record by ID with caching
func (r *DeviceStatusRepository) GetByID(id string) (*entities.DeviceStatus, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("device_status:%s", id)
	cachedData, err := r.cache.Get(cacheKey)
	if err == nil && cachedData != nil {
		var status entities.DeviceStatus
		if err := json.Unmarshal(cachedData, &status); err == nil {
			utils.LogDebug("DeviceStatusRepository: Cache HIT for status ID %s", id)
			return &status, nil
		}
		utils.LogWarn("DeviceStatusRepository: Cache corrupted for status ID %s", id)
	}

	// Cache miss - fetch from database
	utils.LogDebug("DeviceStatusRepository: Cache MISS for status ID %s", id)
	var status entities.DeviceStatus
	err = r.db.Where("id = ?", id).First(&status).Error
	if err != nil {
		return nil, err
	}

	// Save to cache
	if jsonData, err := json.Marshal(status); err == nil {
		r.cache.Set(cacheKey, jsonData)
		utils.LogDebug("DeviceStatusRepository: Cached status ID %s", id)
	}

	return &status, nil
}

// Update updates an existing device status record and invalidates cache
func (r *DeviceStatusRepository) Update(status *entities.DeviceStatus) error {
	err := r.db.Save(status).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("device_status:%s", status.ID)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("DeviceStatusRepository: Failed to invalidate cache for status ID %s: %v", status.ID, err)
	} else {
		utils.LogDebug("DeviceStatusRepository: Invalidated cache for status ID %s", status.ID)
	}

	return nil
}

// Delete soft deletes a device status record by ID and invalidates cache
func (r *DeviceStatusRepository) Delete(id string) error {
	err := r.db.Delete(&entities.DeviceStatus{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("device_status:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("DeviceStatusRepository: Failed to invalidate cache for status ID %s: %v", id, err)
	} else {
		utils.LogDebug("DeviceStatusRepository: Invalidated cache for status ID %s", id)
	}

	return nil
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
