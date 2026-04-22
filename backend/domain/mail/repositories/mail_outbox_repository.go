package repositories

import (
	"sensio/domain/infrastructure"
	"sensio/domain/mail/entities"

	"gorm.io/gorm"
)

type MailOutboxRepository interface {
	Create(entry *entities.MailOutbox) error
	GetByTaskID(taskID string) (*entities.MailOutbox, error)
	GetPendingEntries(limit int) ([]entities.MailOutbox, error)
	UpdateStatus(taskID string, status entities.MailOutboxStatus, errorMsg string) error
	IncrementRetry(taskID string) error
}

type mailOutboxRepository struct {
	db *gorm.DB
}

func NewMailOutboxRepository() MailOutboxRepository {
	return &mailOutboxRepository{
		db: infrastructure.DB,
	}
}

func (r *mailOutboxRepository) Create(entry *entities.MailOutbox) error {
	return r.db.Create(entry).Error
}

func (r *mailOutboxRepository) GetByTaskID(taskID string) (*entities.MailOutbox, error) {
	var entry entities.MailOutbox
	if err := r.db.Where("task_id = ?", taskID).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *mailOutboxRepository) GetPendingEntries(limit int) ([]entities.MailOutbox, error) {
	var entries []entities.MailOutbox
	err := r.db.Where("status = ?", entities.MailOutboxStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

func (r *mailOutboxRepository) UpdateStatus(taskID string, status entities.MailOutboxStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	return r.db.Model(&entities.MailOutbox{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

func (r *mailOutboxRepository) IncrementRetry(taskID string) error {
	return r.db.Model(&entities.MailOutbox{}).
		Where("task_id = ?", taskID).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
}