package repositories

import (
	"encoding/json"
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/entities"

	"gorm.io/gorm"
)

// TeraluxRepository handles database operations for Teralux entities
type TeraluxRepository struct {
	db    *gorm.DB
	cache *infrastructure.BadgerService
}

// NewTeraluxRepository creates a new instance of TeraluxRepository
func NewTeraluxRepository(cache *infrastructure.BadgerService) *TeraluxRepository {
	return &TeraluxRepository{
		db:    infrastructure.DB,
		cache: cache,
	}
}

// Create inserts a new teralux record into the database
func (r *TeraluxRepository) Create(teralux *entities.Teralux) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Create(teralux).Error
}

// GetAll retrieves all active (non-deleted) teralux records
func (r *TeraluxRepository) GetAll() ([]entities.Teralux, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var teraluxList []entities.Teralux
	err := r.db.Find(&teraluxList).Error
	return teraluxList, err
}

// GetByID retrieves a single teralux record by ID with caching
func (r *TeraluxRepository) GetByID(id string) (*entities.Teralux, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("teralux:%s", id)
	cachedData, err := r.cache.Get(cacheKey)
	if err == nil && cachedData != nil {
		var teralux entities.Teralux
		if err := json.Unmarshal(cachedData, &teralux); err == nil {
			utils.LogDebug("TeraluxRepository: Cache HIT for teralux ID %s", id)
			return &teralux, nil
		}
		utils.LogWarn("TeraluxRepository: Cache corrupted for teralux ID %s", id)
	}

	// Cache miss - fetch from database
	utils.LogDebug("TeraluxRepository: Cache MISS for teralux ID %s", id)
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var teralux entities.Teralux
	err = r.db.Where("id = ?", id).First(&teralux).Error
	if err != nil {
		return nil, err
	}

	// Save to cache
	if jsonData, err := json.Marshal(teralux); err == nil {
		r.cache.Set(cacheKey, jsonData)
		utils.LogDebug("TeraluxRepository: Cached teralux ID %s", id)
	}

	return &teralux, nil
}

// GetByMacAddress retrieves a single teralux record by MAC address with caching
func (r *TeraluxRepository) GetByMacAddress(macAddress string) (*entities.Teralux, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("teralux:mac:%s", macAddress)
	cachedData, err := r.cache.Get(cacheKey)
	if err == nil && cachedData != nil {
		var teralux entities.Teralux
		if err := json.Unmarshal(cachedData, &teralux); err == nil {
			utils.LogDebug("TeraluxRepository: Cache HIT for MAC %s", macAddress)
			return &teralux, nil
		}
		utils.LogWarn("TeraluxRepository: Cache corrupted for MAC %s", macAddress)
	}

	// Cache miss - fetch from database
	utils.LogDebug("TeraluxRepository: Cache MISS for MAC %s", macAddress)
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var teralux entities.Teralux
	err = r.db.Where("mac_address = ?", macAddress).First(&teralux).Error
	if err != nil {
		return nil, err
	}

	// Save to cache
	if jsonData, err := json.Marshal(teralux); err == nil {
		r.cache.Set(cacheKey, jsonData)
		utils.LogDebug("TeraluxRepository: Cached teralux MAC %s", macAddress)
	}

	return &teralux, nil
}

// Update updates an existing teralux record and invalidates cache
func (r *TeraluxRepository) Update(teralux *entities.Teralux) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	err := r.db.Save(teralux).Error
	if err != nil {
		return err
	}

	// Invalidate ID cache
	cacheKey := fmt.Sprintf("teralux:%s", teralux.ID)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate ID cache for teralux ID %s: %v", teralux.ID, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated ID cache for teralux ID %s", teralux.ID)
	}

	// Invalidate MAC cache
	macCacheKey := fmt.Sprintf("teralux:mac:%s", teralux.MacAddress)
	if err := r.cache.Delete(macCacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate MAC cache for MAC %s: %v", teralux.MacAddress, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated MAC cache for MAC %s", teralux.MacAddress)
	}

	return nil
}

// Delete soft deletes a teralux record by ID and invalidates cache
func (r *TeraluxRepository) Delete(id string) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	// First, get the teralux to retrieve MAC address for cache invalidation
	var teralux entities.Teralux
	if err := r.db.Where("id = ?", id).First(&teralux).Error; err != nil {
		return err
	}

	err := r.db.Delete(&entities.Teralux{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Invalidate ID cache
	cacheKey := fmt.Sprintf("teralux:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate ID cache for teralux ID %s: %v", id, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated ID cache for teralux ID %s", id)
	}

	// Invalidate MAC cache
	macCacheKey := fmt.Sprintf("teralux:mac:%s", teralux.MacAddress)
	if err := r.cache.Delete(macCacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate MAC cache for MAC %s: %v", teralux.MacAddress, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated MAC cache for MAC %s", teralux.MacAddress)
	}

	return nil
}

// InvalidateCache invalidates both ID and MAC cache for a teralux
func (r *TeraluxRepository) InvalidateCache(id string) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	// Get teralux to find MAC address
	var teralux entities.Teralux
	if err := r.db.Where("id = ?", id).First(&teralux).Error; err != nil {
		// If not found in DB, still try to invalidate ID cache
		cacheKey := fmt.Sprintf("teralux:%s", id)
		_ = r.cache.Delete(cacheKey)
		return err
	}

	// Invalidate ID cache
	cacheKey := fmt.Sprintf("teralux:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate ID cache for teralux ID %s: %v", id, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated ID cache for teralux ID %s", id)
	}

	// Invalidate MAC cache
	macCacheKey := fmt.Sprintf("teralux:mac:%s", teralux.MacAddress)
	if err := r.cache.Delete(macCacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate MAC cache for MAC %s: %v", teralux.MacAddress, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated MAC cache for MAC %s", teralux.MacAddress)
	}

	return nil
}
