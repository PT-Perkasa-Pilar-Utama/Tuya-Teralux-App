package entities

import (
	"time"

	"gorm.io/gorm"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

type ScheduledNotification struct {
	ID             string             `gorm:"type:char(36);primaryKey" json:"id"`
	RoomID         string             `gorm:"type:varchar(255);not null;index" json:"room_id"`
	MacAddress     string             `gorm:"type:varchar(255);not null" json:"mac_address"`
	PhoneNumbers   string             `gorm:"type:text;not null" json:"phone_numbers"`
	BookingInfo    string             `gorm:"type:text" json:"booking_info"`
	BookingTimeEnd string             `gorm:"type:varchar(50)" json:"booking_time_end"`
	ScheduledAt    time.Time          `gorm:"not null;index" json:"scheduled_at"`
	Status         NotificationStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	Template       string             `gorm:"type:varchar(50)" json:"template"`
	CreatedAt      time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
}

func (ScheduledNotification) TableName() string {
	return "scheduled_notifications"
}
