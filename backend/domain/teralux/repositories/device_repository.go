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

// DeviceRepository handles database operations for Device entities
type DeviceRepository struct {
	db    *gorm.DB
	cache *persistence.BadgerService
}

// NewDeviceRepository creates a new instance of DeviceRepository
func NewDeviceRepository(cache *persistence.BadgerService) *DeviceRepository {
	return &DeviceRepository{
		db:    infrastructure.DB,
		cache: cache,
	}
}

// Create inserts a new device record into the database
func (r *DeviceRepository) Create(device *entities.Device) error {
	return r.db.Create(device).Error
}

// GetAll retrieves all active (non-deleted) device records
func (r *DeviceRepository) GetAll() ([]entities.Device, error) {
	var devices []entities.Device
	err := r.db.Find(&devices).Error
	return devices, err
}

// GetByTeraluxID retrieves all devices belonging to a specific Teralux unit
func (r *DeviceRepository) GetByTeraluxID(teraluxID string) ([]entities.Device, error) {
	var devices []entities.Device
	err := r.db.Where("teralux_id = ?", teraluxID).Find(&devices).Error
	return devices, err
}

// GetByID retrieves a single device record by ID with caching
func (r *DeviceRepository) GetByID(id string) (*entities.Device, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("device:%s", id)
	cachedData, err := r.cache.Get(cacheKey)
	if err == nil && cachedData != nil {
		var device entities.Device
		if err := json.Unmarshal(cachedData, &device); err == nil {
			utils.LogDebug("DeviceRepository: Cache HIT for device ID %s", id)
			return &device, nil
		}
		utils.LogWarn("DeviceRepository: Cache corrupted for device ID %s", id)
	}

	// Cache miss - fetch from database
	utils.LogDebug("DeviceRepository: Cache MISS for device ID %s", id)
	var device entities.Device
	err = r.db.Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}

	// Save to cache
	if jsonData, err := json.Marshal(device); err == nil {
		r.cache.Set(cacheKey, jsonData)
		utils.LogDebug("DeviceRepository: Cached device ID %s", id)
	}

	return &device, nil
}

// Update updates an existing device record and invalidates cache
func (r *DeviceRepository) Update(device *entities.Device) error {
	err := r.db.Save(device).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("device:%s", device.ID)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("DeviceRepository: Failed to invalidate cache for device ID %s: %v", device.ID, err)
	} else {
		utils.LogDebug("DeviceRepository: Invalidated cache for device ID %s", device.ID)
	}

	return nil
}

// Delete soft deletes a device record by ID and invalidates cache
func (r *DeviceRepository) Delete(id string) error {
	err := r.db.Delete(&entities.Device{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("device:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("DeviceRepository: Failed to invalidate cache for device ID %s: %v", id, err)
	} else {
		utils.LogDebug("DeviceRepository: Invalidated cache for device ID %s", id)
	}

	return nil
}

// GetByRemoteID retrieves a device by its Remote ID
func (r *DeviceRepository) GetByRemoteID(remoteID string) (*entities.Device, error) {
	var device entities.Device
	err := r.db.Where("remote_id = ?", remoteID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}
