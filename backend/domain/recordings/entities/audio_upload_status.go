package entities

import (
	"time"
)

// AudioUploadStatus represents the status of an audio file upload to S3
type AudioUploadStatus struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"column:user_id" json:"user_id"`
	ObjectKey string    `gorm:"column:object_key" json:"object_key"`
	Status    string    `gorm:"column:status;type:varchar(20)" json:"status"` // PENDING, COMPLETED, FAILED
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	AudioUploadStatusPending   = "PENDING"
	AudioUploadStatusCompleted = "COMPLETED"
	AudioUploadStatusFailed    = "FAILED"
)

func (AudioUploadStatus) TableName() string {
	return "audio_upload_status"
}