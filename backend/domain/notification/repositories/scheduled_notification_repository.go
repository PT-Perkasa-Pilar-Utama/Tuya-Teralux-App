package notification

import (
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	notificationEntities "sensio/domain/notification/entities"

	"gorm.io/gorm"
)

type IScheduledNotificationRepository interface {
	Create(notification *notificationEntities.ScheduledNotification) error
	GetByID(id string) (*notificationEntities.ScheduledNotification, error)
	GetPendingNotifications(limit int) ([]notificationEntities.ScheduledNotification, error)
	UpdateStatus(id string, status notificationEntities.NotificationStatus) error
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

func (r *ScheduledNotificationRepository) Create(notification *notificationEntities.ScheduledNotification) error {
	if r.db == nil {
		return utils.NewAPIError(500, "database not initialized")
	}
	return r.db.Create(notification).Error
}

func (r *ScheduledNotificationRepository) GetByID(id string) (*notificationEntities.ScheduledNotification, error) {
	if r.db == nil {
		return nil, utils.NewAPIError(500, "database not initialized")
	}
	var notification notificationEntities.ScheduledNotification
	err := r.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *ScheduledNotificationRepository) GetPendingNotifications(limit int) ([]notificationEntities.ScheduledNotification, error) {
	if r.db == nil {
		return nil, utils.NewAPIError(500, "database not initialized")
	}
	var notifications []notificationEntities.ScheduledNotification
	err := r.db.Where("status = ?", notificationEntities.NotificationStatusPending).
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *ScheduledNotificationRepository) UpdateStatus(id string, status notificationEntities.NotificationStatus) error {
	if r.db == nil {
		return utils.NewAPIError(500, "database not initialized")
	}
	return r.db.Model(&notificationEntities.ScheduledNotification{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *ScheduledNotificationRepository) Delete(id string) error {
	if r.db == nil {
		return utils.NewAPIError(500, "database not initialized")
	}
	return r.db.Delete(&notificationEntities.ScheduledNotification{}, "id = ?", id).Error
}
