package entities

import (
	"time"

	"gorm.io/gorm"
)

// DeviceStatus represents a status/attribute of a device
type DeviceStatus struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	DeviceID  string         `gorm:"type:char(36);not null;index" json:"device_id"`
	Code      string         `gorm:"type:varchar(255);not null;index" json:"code"`
	Value     string         `gorm:"type:text" json:"value"` // JSON-encoded value from Tuya
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for the DeviceStatus model
func (DeviceStatus) TableName() string {
	return "device_statuses"
}
