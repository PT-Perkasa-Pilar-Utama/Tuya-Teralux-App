package entities

import (
	"time"

	"gorm.io/gorm"
)

type DeviceStatus struct {
	DeviceID  string `gorm:"type:char(36);primaryKey"`
	Code      string `gorm:"type:varchar(255);primaryKey"`
	Value     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (DeviceStatus) TableName() string {
	return "device_statuses"
}
