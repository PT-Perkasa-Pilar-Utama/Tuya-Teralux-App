package entities

import (
	"time"
)

// Recording represents an audio recording file
type Recording struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	Filename      string    `json:"filename"`        // UUIDv4 filename (e.g., 123e4567-e89b-12d3-a456-426614174000.mp3)
	OriginalName  string    `json:"original_name"`   // Original filename uploaded by user
	AudioUrl      string    `json:"audio_url"`      // Public URL to access the file
	MacAddress    string    `json:"mac_address"`    // Device MAC Address
	S3Key         string    `gorm:"column:s3_key" json:"s3_key"`
	BookingID     string    `gorm:"column:booking_id" json:"booking_id"`
	PasswordHash  string    `gorm:"column:password_hash" json:"password_hash"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Recording) TableName() string {
	return "recordings"
}
