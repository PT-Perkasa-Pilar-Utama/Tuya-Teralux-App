package entities

import (
	"time"

	"gorm.io/gorm"
)

// Terminal represents a terminal device in the system
type Terminal struct {
	ID           string         `gorm:"type:char(36);primaryKey" json:"id"`
	MacAddress   string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"mac_address"`
	RoomID       string         `gorm:"type:varchar(255);not null" json:"room_id"`
	TuyaUID      string         `gorm:"type:varchar(255);index" json:"tuya_uid"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	DeviceTypeID string         `gorm:"type:varchar(255)" json:"device_type_id"`
	AiProvider   *string        `gorm:"type:varchar(50);index" json:"ai_provider"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for the Terminal model
func (Terminal) TableName() string {
	return "terminal"
}
