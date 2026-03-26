package repositories

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/recordings/entities"
)

type RecordingRepository interface {
	Save(recording *entities.Recording) error
	GetAll(page, limit int) ([]entities.Recording, int64, error)
	GetByID(id string) (*entities.Recording, error)
	Delete(id string) error
}

type recordingRepository struct {
	db     *gorm.DB
	badger *infrastructure.BadgerService // Kept for consistency, though metadata is in SQL
}

func NewRecordingRepository(badger *infrastructure.BadgerService) RecordingRepository {
	return &recordingRepository{
		db:     infrastructure.DB,
		badger: badger,
	}
}

func (r *recordingRepository) Save(recording *entities.Recording) error {
	start := time.Now()
	result := r.db.Create(recording)
	duration := time.Since(start)

	if result.Error != nil {
		log.Printf("RecordingRepository: Save failed | duration_ms=%d | error=%v", duration.Milliseconds(), result.Error)
		return result.Error
	}

	log.Printf("RecordingRepository: Save completed | id=%s | duration_ms=%d | rows=%d", recording.ID, duration.Milliseconds(), result.RowsAffected)
	return nil
}

func (r *recordingRepository) GetAll(page, limit int) ([]entities.Recording, int64, error) {
	start := time.Now()
	var recordings []entities.Recording
	var total int64

	offset := (page - 1) * limit

	// Count total records
	countStart := time.Now()
	if err := r.db.Model(&entities.Recording{}).Count(&total).Error; err != nil {
		log.Printf("RecordingRepository: GetAll count failed | duration_ms=%d | error=%v", time.Since(countStart).Milliseconds(), err)
		return nil, 0, err
	}
	utils.LogDebug("RecordingRepository: GetAll count completed | total=%d | duration_ms=%d", total, time.Since(countStart).Milliseconds())

	// Get paginated records
	queryStart := time.Now()
	result := r.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&recordings)
	queryDuration := time.Since(queryStart)

	if result.Error != nil {
		log.Printf("RecordingRepository: GetAll query failed | page=%d | limit=%d | duration_ms=%d | error=%v", page, limit, queryDuration.Milliseconds(), result.Error)
		return nil, 0, result.Error
	}

	totalDuration := time.Since(start)
	utils.LogDebug("RecordingRepository: GetAll completed | page=%d | limit=%d | returned=%d | total=%d | duration_ms=%d", page, limit, len(recordings), total, totalDuration.Milliseconds())
	return recordings, total, nil
}

func (r *recordingRepository) GetByID(id string) (*entities.Recording, error) {
	start := time.Now()

	// Try to get from cache first
	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("recording:%s", id)
	cachedData, err := r.badger.Get(cacheKey)
	cacheDuration := time.Since(cacheStart)

	if err == nil && cachedData != nil {
		var recording entities.Recording
		if err := json.Unmarshal(cachedData, &recording); err == nil {
			utils.LogDebug("RecordingRepository: Cache HIT for recording ID %s | cache_duration_ms=%d | total_duration_ms=%d", id, cacheDuration.Milliseconds(), time.Since(start).Milliseconds())
			return &recording, nil
		}
		utils.LogWarn("RecordingRepository: Cache corrupted for recording ID %s | unmarshal_error=%v", id, err)
	}

	// Cache miss - fetch from database
	utils.LogDebug("RecordingRepository: Cache MISS for recording ID %s | cache_duration_ms=%d", id, cacheDuration.Milliseconds())

	dbStart := time.Now()
	var recording entities.Recording
	if err := r.db.First(&recording, "id = ?", id).Error; err != nil {
		utils.LogDebug("RecordingRepository: Database query failed for recording ID %s | db_duration_ms=%d | error=%v", id, time.Since(dbStart).Milliseconds(), err)
		return nil, err
	}
	dbDuration := time.Since(dbStart)
	utils.LogDebug("RecordingRepository: Database query completed for recording ID %s | db_duration_ms=%d", id, dbDuration.Milliseconds())

	// Save to cache
	if jsonData, err := json.Marshal(recording); err == nil {
		cacheSetStart := time.Now()
		if err := r.badger.Set(cacheKey, jsonData); err != nil {
			utils.LogWarn("RecordingRepository: Failed to cache recording ID %s | cache_set_duration_ms=%d | error=%v", id, time.Since(cacheSetStart).Milliseconds(), err)
		} else {
			utils.LogDebug("RecordingRepository: Cached recording ID %s | cache_set_duration_ms=%d", id, time.Since(cacheSetStart).Milliseconds())
		}
	}

	totalDuration := time.Since(start)
	utils.LogDebug("RecordingRepository: GetByID completed for recording ID %s | total_duration_ms=%d", id, totalDuration.Milliseconds())
	return &recording, nil
}

func (r *recordingRepository) Delete(id string) error {
	start := time.Now()

	dbStart := time.Now()
	result := r.db.Delete(&entities.Recording{}, "id = ?", id)
	dbDuration := time.Since(dbStart)

	if result.Error != nil {
		log.Printf("RecordingRepository: Delete failed | id=%s | db_duration_ms=%d | error=%v", id, dbDuration.Milliseconds(), result.Error)
		return result.Error
	}
	utils.LogDebug("RecordingRepository: Delete completed | id=%s | db_duration_ms=%d", id, dbDuration.Milliseconds())

	// Invalidate cache
	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("recording:%s", id)
	if err := r.badger.Delete(cacheKey); err != nil {
		utils.LogWarn("RecordingRepository: Failed to invalidate cache for recording ID %s | cache_duration_ms=%d | error=%v", id, time.Since(cacheStart).Milliseconds(), err)
	} else {
		utils.LogDebug("RecordingRepository: Invalidated cache for recording ID %s | cache_duration_ms=%d", id, time.Since(cacheStart).Milliseconds())
	}

	totalDuration := time.Since(start)
	utils.LogDebug("RecordingRepository: Delete completed | id=%s | total_duration_ms=%d", id, totalDuration.Milliseconds())
	return nil
}
