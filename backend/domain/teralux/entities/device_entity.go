package entities

import (
	"time"

	"gorm.io/gorm"
)

// Device represents a device connected to a teralux unit
type Device struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	TeraluxID string         `gorm:"type:char(36);not null;index" json:"teralux_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Online    bool           `gorm:"not null;default:false" json:"online"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Associations can be added here if needed, e.g., Teralux *Teralux
}

// TableName specifies the table name for the Device model
func (Device) TableName() string {
	return "devices"
}
