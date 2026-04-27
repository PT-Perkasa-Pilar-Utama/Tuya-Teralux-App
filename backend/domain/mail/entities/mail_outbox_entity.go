package entities

import (
	"time"

	"gorm.io/gorm"
)

// MailOutboxStatus represents the processing status of a mail outbox entry
type MailOutboxStatus string

const (
	MailOutboxStatusPending MailOutboxStatus = "pending"
	MailOutboxStatusSent    MailOutboxStatus = "sent"
	MailOutboxStatusFailed  MailOutboxStatus = "failed"
)

// MailOutbox represents a queued mail job for reliable delivery
// Implements fire-and-forget pattern with persistence guarantee
type MailOutbox struct {
	ID           uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID       string           `gorm:"type:varchar(36);uniqueIndex;not null" json:"task_id"`
	Recipient    string           `gorm:"type:varchar(255);not null" json:"recipient"`
	ObjectKey    string           `gorm:"type:varchar(512);not null" json:"object_key"`
	Purpose      string           `gorm:"type:varchar(255);not null" json:"purpose"`
	Subject      string           `gorm:"type:varchar(500);not null" json:"subject"`
	Status       MailOutboxStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	RetryCount   int              `gorm:"type:int;not null;default:0" json:"retry_count"`
	ErrorMessage string           `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time        `gorm:"not null;index" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for MailOutbox
func (MailOutbox) TableName() string {
	return "mail_outbox"
}
