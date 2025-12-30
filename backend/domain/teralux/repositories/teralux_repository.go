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

// TeraluxRepository handles database operations for Teralux entities
type TeraluxRepository struct {
	db    *gorm.DB
	cache *persistence.BadgerService
}

// NewTeraluxRepository creates a new instance of TeraluxRepository
func NewTeraluxRepository(cache *persistence.BadgerService) *TeraluxRepository {
	return &TeraluxRepository{
		db:    infrastructure.DB,
		cache: cache,
	}
}

// Create inserts a new teralux record into the database
func (r *TeraluxRepository) Create(teralux *entities.Teralux) error {
	return r.db.Create(teralux).Error
}

// GetAll retrieves all active (non-deleted) teralux records
func (r *TeraluxRepository) GetAll() ([]entities.Teralux, error) {
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

// Update updates an existing teralux record and invalidates cache
func (r *TeraluxRepository) Update(teralux *entities.Teralux) error {
	err := r.db.Save(teralux).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("teralux:%s", teralux.ID)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate cache for teralux ID %s: %v", teralux.ID, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated cache for teralux ID %s", teralux.ID)
	}

	return nil
}

// Delete soft deletes a teralux record by ID and invalidates cache
func (r *TeraluxRepository) Delete(id string) error {
	err := r.db.Delete(&entities.Teralux{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("teralux:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TeraluxRepository: Failed to invalidate cache for teralux ID %s: %v", id, err)
	} else {
		utils.LogDebug("TeraluxRepository: Invalidated cache for teralux ID %s", id)
	}

	return nil
}
