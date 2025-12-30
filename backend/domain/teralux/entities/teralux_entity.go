package entities

import (
	"time"

	"gorm.io/gorm"
)

// Teralux represents a teralux device in the system
type Teralux struct {
	ID         string         `gorm:"type:char(36);primaryKey" json:"id"`
	MacAddress string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"mac_address"`
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for the Teralux model
func (Teralux) TableName() string {
	return "teralux"
}
