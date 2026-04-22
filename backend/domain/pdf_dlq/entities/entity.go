package entities

import (
	"time"
)

// PDFDeadLetter represents a failed PDF generation/upload job entry
type PDFDeadLetter struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	JobID         string    `gorm:"column:job_id;index" json:"job_id"`
	Status        string    `gorm:"column:status;default:'PENDING'" json:"status"`
	FailureReason string    `gorm:"column:failure_reason;type:text" json:"failure_reason"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	RetryCount    int       `gorm:"column:retry_count;default:0" json:"retry_count"`
	LastRetryAt   *time.Time `gorm:"column:last_retry_at" json:"last_retry_at,omitempty"`
}

func (PDFDeadLetter) TableName() string {
	return "pdf_dead_letter"
}