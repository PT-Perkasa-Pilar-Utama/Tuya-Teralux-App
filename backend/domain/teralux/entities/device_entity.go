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

	// Tuya Fields
	TuyaID            string `gorm:"index" json:"tuya_id"`
	RemoteID          string `gorm:"index" json:"remote_id"`
	Category          string `json:"category"`
	RemoteCategory    string `json:"remote_category"`
	ProductName       string `json:"product_name"`
	RemoteProductName string `json:"remote_product_name"`
	LocalKey          string `json:"local_key"`
	GatewayID         string `json:"gateway_id"`
	IP                string `json:"ip"`
	Model             string `json:"model"`
	Icon              string `json:"icon"`
}

// TableName specifies the table name for the Device model
func (Device) TableName() string {
	return "devices"
}
