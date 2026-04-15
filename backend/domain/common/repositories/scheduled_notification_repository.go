package repositories

import (
	"sensio/domain/common/entities"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"

	"gorm.io/gorm"
)

type IScheduledNotificationRepository interface {
	Create(notification *entities.ScheduledNotification) error
	GetByID(id string) (*entities.ScheduledNotification, error)
	GetPendingNotifications(limit int) ([]entities.ScheduledNotification, error)
	UpdateStatus(id string, status entities.NotificationStatus) error
	Delete(id string) error
}

type ScheduledNotificationRepository struct {
	db *gorm.DB
}

func NewScheduledNotificationRepository() *ScheduledNotificationRepository {
	return &ScheduledNotificationRepository{
		db: infrastructure.DB,
	}
}

func (r *ScheduledNotificationRepository) Create(notification *entities.ScheduledNotification) error {
	if r.db == nil {
		return utils.NewAPIError(500, "database not initialized")
	}
	return r.db.Create(notification).Error
}

func (r *ScheduledNotificationRepository) GetByID(id string) (*entities.ScheduledNotification, error) {
	if r.db == nil {
		return nil, utils.NewAPIError(500, "database not initialized")
	}
	var notification entities.ScheduledNotification
	err := r.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *ScheduledNotificationRepository) GetPendingNotifications(limit int) ([]entities.ScheduledNotification, error) {
	if r.db == nil {
		return nil, utils.NewAPIError(500, "database not initialized")
	}
	var notifications []entities.ScheduledNotification
	err := r.db.Where("status = ?", entities.NotificationStatusPending).
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *ScheduledNotificationRepository) UpdateStatus(id string, status entities.NotificationStatus) error {
	if r.db == nil {
		return utils.NewAPIError(500, "database not initialized")
	}
	return r.db.Model(&entities.ScheduledNotification{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *ScheduledNotificationRepository) Delete(id string) error {
	if r.db == nil {
		return utils.NewAPIError(500, "database not initialized")
	}
	return r.db.Delete(&entities.ScheduledNotification{}, "id = ?", id).Error
}
