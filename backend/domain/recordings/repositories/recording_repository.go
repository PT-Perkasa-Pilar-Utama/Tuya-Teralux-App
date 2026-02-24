package repositories

import (
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/recordings/entities"
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
	result := r.db.Create(recording)
	if result.Error != nil {
		log.Printf("Error saving recording: %v", result.Error)
		return result.Error
	}
	return nil
}

func (r *recordingRepository) GetAll(page, limit int) ([]entities.Recording, int64, error) {
	var recordings []entities.Recording
	var total int64

	offset := (page - 1) * limit

	// Count total records
	if err := r.db.Model(&entities.Recording{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	result := r.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&recordings)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return recordings, total, nil
}

func (r *recordingRepository) GetByID(id string) (*entities.Recording, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("recording:%s", id)
	cachedData, err := r.badger.Get(cacheKey)
	if err == nil && cachedData != nil {
		var recording entities.Recording
		if err := json.Unmarshal(cachedData, &recording); err == nil {
			utils.LogDebug("RecordingRepository: Cache HIT for recording ID %s", id)
			return &recording, nil
		}
		utils.LogWarn("RecordingRepository: Cache corrupted for recording ID %s", id)
	}

	// Cache miss - fetch from database
	utils.LogDebug("RecordingRepository: Cache MISS for recording ID %s", id)
	var recording entities.Recording
	if err := r.db.First(&recording, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Save to cache
	if jsonData, err := json.Marshal(recording); err == nil {
		if err := r.badger.Set(cacheKey, jsonData); err != nil {
			utils.LogWarn("RecordingRepository: Failed to cache recording ID %s: %v", id, err)
		} else {
			utils.LogDebug("RecordingRepository: Cached recording ID %s", id)
		}
	}

	return &recording, nil
}

func (r *recordingRepository) Delete(id string) error {
	result := r.db.Delete(&entities.Recording{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("recording:%s", id)
	if err := r.badger.Delete(cacheKey); err != nil {
		utils.LogWarn("RecordingRepository: Failed to invalidate cache for recording ID %s: %v", id, err)
	} else {
		utils.LogDebug("RecordingRepository: Invalidated cache for recording ID %s", id)
	}

	return nil
}
