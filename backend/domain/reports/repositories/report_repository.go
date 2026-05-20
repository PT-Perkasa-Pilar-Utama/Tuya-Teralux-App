package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/reports/entities"
)

type ReportRepository interface {
	Save(report *entities.Report) error
	GetByID(id string) (*entities.Report, error)
	Delete(id string) error
}

type reportRepository struct {
	db     *gorm.DB
	badger *infrastructure.BadgerService
}

func NewReportRepository(badger *infrastructure.BadgerService) ReportRepository {
	return &reportRepository{
		db:     infrastructure.DB,
		badger: badger,
	}
}

func (r *reportRepository) Save(report *entities.Report) error {
	start := time.Now()
	result := r.db.Create(report)
	duration := time.Since(start)

	if result.Error != nil {
		utils.LogError("ReportRepository: Save failed | id=%s | duration_ms=%d | error=%v", report.ID, duration.Milliseconds(), result.Error)
		return result.Error
	}

	utils.LogDebug("ReportRepository: Save completed | id=%s | duration_ms=%d | rows=%d", report.ID, duration.Milliseconds(), result.RowsAffected)
	return nil
}

func (r *reportRepository) GetByID(id string) (*entities.Report, error) {
	start := time.Now()
	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("report:%s", id)
	cachedData, err := r.badger.Get(cacheKey)
	cacheDuration := time.Since(cacheStart)

	if err == nil && cachedData != nil {
		var report entities.Report
		if err := json.Unmarshal(cachedData, &report); err == nil {
			utils.LogDebug("ReportRepository: Cache HIT for report ID %s | cache_duration_ms=%d | total_duration_ms=%d", id, cacheDuration.Milliseconds(), time.Since(start).Milliseconds())
			return &report, nil
		}
		utils.LogWarn("ReportRepository: Cache corrupted for report ID %s | unmarshal_error=%v", id, err)
	}

	utils.LogDebug("ReportRepository: Cache MISS for report ID %s | cache_duration_ms=%d", id, cacheDuration.Milliseconds())

	dbStart := time.Now()
	var report entities.Report
	if err := r.db.First(&report, "id = ?", id).Error; err != nil {
		utils.LogDebug("ReportRepository: Database query failed for report ID %s | db_duration_ms=%d | error=%v", id, time.Since(dbStart).Milliseconds(), err)
		return nil, err
	}
	dbDuration := time.Since(dbStart)
	utils.LogDebug("ReportRepository: Database query completed for report ID %s | db_duration_ms=%d", id, dbDuration.Milliseconds())

	if jsonData, err := json.Marshal(report); err == nil {
		cacheSetStart := time.Now()
		if err := r.badger.Set(cacheKey, jsonData); err != nil {
			utils.LogWarn("ReportRepository: Failed to cache report ID %s | cache_set_duration_ms=%d | error=%v", id, time.Since(cacheSetStart).Milliseconds(), err)
		} else {
			utils.LogDebug("ReportRepository: Cached report ID %s | cache_set_duration_ms=%d", id, time.Since(cacheSetStart).Milliseconds())
		}
	}

	totalDuration := time.Since(start)
	utils.LogDebug("ReportRepository: GetByID completed for report ID %s | total_duration_ms=%d", id, totalDuration.Milliseconds())
	return &report, nil
}

func (r *reportRepository) Delete(id string) error {
	start := time.Now()
	dbStart := time.Now()
	result := r.db.Delete(&entities.Report{}, "id = ?", id)
	dbDuration := time.Since(dbStart)

	if result.Error != nil {
		utils.LogError("ReportRepository: Delete failed | id=%s | db_duration_ms=%d | error=%v", id, dbDuration.Milliseconds(), result.Error)
		return result.Error
	}
	utils.LogDebug("ReportRepository: Delete completed | id=%s | db_duration_ms=%d", id, dbDuration.Milliseconds())

	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("report:%s", id)
	if err := r.badger.Delete(cacheKey); err != nil {
		utils.LogWarn("ReportRepository: Failed to invalidate cache for report ID %s | cache_duration_ms=%d | error=%v", id, time.Since(cacheStart).Milliseconds(), err)
	} else {
		utils.LogDebug("ReportRepository: Invalidated cache for report ID %s | cache_duration_ms=%d", id, time.Since(cacheStart).Milliseconds())
	}

	totalDuration := time.Since(start)
	utils.LogDebug("ReportRepository: Delete completed | id=%s | total_duration_ms=%d", id, totalDuration.Milliseconds())
	return nil
}
