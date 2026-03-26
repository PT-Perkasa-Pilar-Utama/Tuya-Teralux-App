package repositories

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/entities"
	"time"

	"gorm.io/gorm"
)

// ITerminalRepository defines the interface for terminal storage operations
type ITerminalRepository interface {
	Create(terminal *entities.Terminal) error
	GetAll() ([]entities.Terminal, error)
	GetAllPaginated(offset, limit int, roomID *string) ([]entities.Terminal, int64, error)
	GetByID(id string) (*entities.Terminal, error)
	GetByMacAddress(macAddress string) (*entities.Terminal, error)
	GetByRoomID(roomID string) ([]entities.Terminal, error)
	Update(terminal *entities.Terminal) error
	Delete(id string) error
	InvalidateCache(id string) error

	// MQTT User methods
	CreateMQTTUser(user *entities.MQTTUser) error
	GetMQTTUserByUsername(username string) (*entities.MQTTUser, error)
}

// TerminalRepository handles database operations for Terminal entities
type TerminalRepository struct {
	db    *gorm.DB
	cache *infrastructure.BadgerService
}

// NewTerminalRepository creates a new instance of TerminalRepository
func NewTerminalRepository(cache *infrastructure.BadgerService) *TerminalRepository {
	return &TerminalRepository{
		db:    infrastructure.DB,
		cache: cache,
	}
}

// Create inserts a new terminal record into the database
func (r *TerminalRepository) Create(terminal *entities.Terminal) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Create(terminal).Error
}

// GetAll retrieves all active (non-deleted) terminal records
func (r *TerminalRepository) GetAll() ([]entities.Terminal, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var terminalList []entities.Terminal
	err := r.db.Find(&terminalList).Error
	return terminalList, err
}

// GetAllPaginated retrieves terminal records with pagination and optional room filtering
func (r *TerminalRepository) GetAllPaginated(offset, limit int, roomID *string) ([]entities.Terminal, int64, error) {
	if r.db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}
	var terminalList []entities.Terminal
	var total int64

	query := r.db.Model(&entities.Terminal{})

	// Apply filter
	if roomID != nil {
		query = query.Where("room_id = ?", *roomID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	err := query.Find(&terminalList).Error
	return terminalList, total, err
}

// GetByID retrieves a single terminal record by ID with caching
func (r *TerminalRepository) GetByID(id string) (*entities.Terminal, error) {
	start := time.Now()

	// Try to get from cache first
	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("terminal:%s", id)
	cachedData, err := r.cache.Get(cacheKey)
	cacheDuration := time.Since(cacheStart)

	if err == nil && cachedData != nil {
		var terminal entities.Terminal
		if err := json.Unmarshal(cachedData, &terminal); err == nil {
			utils.LogDebug("TerminalRepository: Cache HIT for terminal ID %s | cache_duration_ms=%d | total_duration_ms=%d", id, cacheDuration.Milliseconds(), time.Since(start).Milliseconds())
			return &terminal, nil
		}
		utils.LogWarn("TerminalRepository: Cache corrupted for terminal ID %s | unmarshal_error=%v", id, err)
	}

	// Cache miss - fetch from database
	utils.LogDebug("TerminalRepository: Cache MISS for terminal ID %s | cache_duration_ms=%d", id, cacheDuration.Milliseconds())

	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbStart := time.Now()
	var terminal entities.Terminal
	err = r.db.Where("id = ?", id).First(&terminal).Error
	dbDuration := time.Since(dbStart)

	if err != nil {
		utils.LogDebug("TerminalRepository: Database query failed for terminal ID %s | db_duration_ms=%d | error=%v", id, dbDuration.Milliseconds(), err)
		return nil, err
	}
	utils.LogDebug("TerminalRepository: Database query completed for terminal ID %s | db_duration_ms=%d | rows=1", id, dbDuration.Milliseconds())

	// Save to cache
	if jsonData, err := json.Marshal(terminal); err == nil {
		cacheSetStart := time.Now()
		if err := r.cache.Set(cacheKey, jsonData); err != nil {
			utils.LogWarn("TerminalRepository: Failed to cache terminal ID %s | cache_set_duration_ms=%d | error=%v", id, time.Since(cacheSetStart).Milliseconds(), err)
		} else {
			utils.LogDebug("TerminalRepository: Cached terminal ID %s | cache_set_duration_ms=%d", id, time.Since(cacheSetStart).Milliseconds())
		}
	}

	totalDuration := time.Since(start)
	utils.LogDebug("TerminalRepository: GetByID completed for terminal ID %s | total_duration_ms=%d", id, totalDuration.Milliseconds())
	return &terminal, nil
}

// GetByMacAddress retrieves a single terminal record by MAC address with caching
func (r *TerminalRepository) GetByMacAddress(macAddress string) (*entities.Terminal, error) {
	start := time.Now()

	// Try to get from cache first
	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("terminal:mac:%s", macAddress)
	cachedData, err := r.cache.Get(cacheKey)
	cacheDuration := time.Since(cacheStart)

	if err == nil && cachedData != nil {
		var terminal entities.Terminal
		if err := json.Unmarshal(cachedData, &terminal); err == nil {
			utils.LogDebug("TerminalRepository: Cache HIT for MAC %s | cache_duration_ms=%d | total_duration_ms=%d", macAddress, cacheDuration.Milliseconds(), time.Since(start).Milliseconds())
			return &terminal, nil
		}
		utils.LogWarn("TerminalRepository: Cache corrupted for MAC %s | unmarshal_error=%v", macAddress, err)
	}

	// Cache miss - fetch from database
	utils.LogDebug("TerminalRepository: Cache MISS for MAC %s | cache_duration_ms=%d", macAddress, cacheDuration.Milliseconds())

	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbStart := time.Now()
	var terminal entities.Terminal
	err = r.db.Where("mac_address = ?", macAddress).First(&terminal).Error
	dbDuration := time.Since(dbStart)

	if err != nil {
		utils.LogDebug("TerminalRepository: Database query failed for MAC %s | db_duration_ms=%d | error=%v", macAddress, dbDuration.Milliseconds(), err)
		return nil, err
	}
	utils.LogDebug("TerminalRepository: Database query completed for MAC %s | db_duration_ms=%d | rows=1", macAddress, dbDuration.Milliseconds())

	// Save to cache
	if jsonData, err := json.Marshal(terminal); err == nil {
		cacheSetStart := time.Now()
		if err := r.cache.Set(cacheKey, jsonData); err != nil {
			utils.LogWarn("TerminalRepository: Failed to cache terminal MAC %s | cache_set_duration_ms=%d | error=%v", macAddress, time.Since(cacheSetStart).Milliseconds(), err)
		} else {
			utils.LogDebug("TerminalRepository: Cached terminal MAC %s | cache_set_duration_ms=%d", macAddress, time.Since(cacheSetStart).Milliseconds())
		}
	}

	totalDuration := time.Since(start)
	utils.LogDebug("TerminalRepository: GetByMacAddress completed for MAC %s | total_duration_ms=%d", macAddress, totalDuration.Milliseconds())
	return &terminal, nil
}

// GetByRoomID retrieves all active (non-deleted) terminal records for a specific room ID
func (r *TerminalRepository) GetByRoomID(roomID string) ([]entities.Terminal, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var terminalList []entities.Terminal
	err := r.db.Where("room_id = ?", roomID).Find(&terminalList).Error
	return terminalList, err
}

// Update updates an existing terminal record and invalidates cache
func (r *TerminalRepository) Update(terminal *entities.Terminal) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	err := r.db.Save(terminal).Error
	if err != nil {
		return err
	}

	// Invalidate ID cache
	cacheKey := fmt.Sprintf("terminal:%s", terminal.ID)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TerminalRepository: Failed to invalidate ID cache for terminal ID %s: %v", terminal.ID, err)
	} else {
		utils.LogDebug("TerminalRepository: Invalidated ID cache for terminal ID %s", terminal.ID)
	}

	// Invalidate MAC cache
	macCacheKey := fmt.Sprintf("terminal:mac:%s", terminal.MacAddress)
	if err := r.cache.Delete(macCacheKey); err != nil {
		utils.LogWarn("TerminalRepository: Failed to invalidate MAC cache for MAC %s: %v", terminal.MacAddress, err)
	} else {
		utils.LogDebug("TerminalRepository: Invalidated MAC cache for MAC %s", terminal.MacAddress)
	}

	return nil
}

// Delete soft deletes a terminal record by ID and invalidates cache
func (r *TerminalRepository) Delete(id string) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	// First, get the terminal to retrieve MAC address for cache invalidation
	var terminal entities.Terminal
	if err := r.db.Where("id = ?", id).First(&terminal).Error; err != nil {
		return err
	}

	err := r.db.Delete(&entities.Terminal{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Invalidate ID cache
	cacheKey := fmt.Sprintf("terminal:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TerminalRepository: Failed to invalidate ID cache for terminal ID %s: %v", id, err)
	} else {
		utils.LogDebug("TerminalRepository: Invalidated ID cache for terminal ID %s", id)
	}

	// Invalidate MAC cache
	macCacheKey := fmt.Sprintf("terminal:mac:%s", terminal.MacAddress)
	if err := r.cache.Delete(macCacheKey); err != nil {
		utils.LogWarn("TerminalRepository: Failed to invalidate MAC cache for MAC %s: %v", terminal.MacAddress, err)
	} else {
		utils.LogDebug("TerminalRepository: Invalidated MAC cache for MAC %s", terminal.MacAddress)
	}

	return nil
}

// InvalidateCache invalidates both ID and MAC cache for a terminal
func (r *TerminalRepository) InvalidateCache(id string) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	// Get terminal to find MAC address
	var terminal entities.Terminal
	if err := r.db.Where("id = ?", id).First(&terminal).Error; err != nil {
		// If not found in DB, still try to invalidate ID cache
		cacheKey := fmt.Sprintf("terminal:%s", id)
		_ = r.cache.Delete(cacheKey)
		return err
	}

	// Invalidate ID cache
	cacheKey := fmt.Sprintf("terminal:%s", id)
	if err := r.cache.Delete(cacheKey); err != nil {
		utils.LogWarn("TerminalRepository: Failed to invalidate ID cache for terminal ID %s: %v", id, err)
	} else {
		utils.LogDebug("TerminalRepository: Invalidated ID cache for terminal ID %s", id)
	}

	// Invalidate MAC cache
	macCacheKey := fmt.Sprintf("terminal:mac:%s", terminal.MacAddress)
	if err := r.cache.Delete(macCacheKey); err != nil {
		utils.LogWarn("TerminalRepository: Failed to invalidate MAC cache for MAC %s: %v", terminal.MacAddress, err)
	} else {
		utils.LogDebug("TerminalRepository: Invalidated MAC cache for MAC %s", terminal.MacAddress)
	}

	return nil
}

// CreateMQTTUser inserts a new MQTT user into the mqtt_users table
func (r *TerminalRepository) CreateMQTTUser(user *entities.MQTTUser) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Create(user).Error
}

// GetMQTTUserByUsername retrieves an MQTT user by username
func (r *TerminalRepository) GetMQTTUserByUsername(username string) (*entities.MQTTUser, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var user entities.MQTTUser
	err := r.db.Where("username = ? AND is_deleted = false", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
