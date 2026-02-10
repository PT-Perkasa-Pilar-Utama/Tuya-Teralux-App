package entities

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Action represents a single instruction within a scene
type Action struct {
	DeviceID string      `json:"device_id,omitempty"`
	Code     string      `json:"code,omitempty"`
	RemoteID string      `json:"remote_id,omitempty"` // For IR devices
	Topic    string      `json:"topic,omitempty"`     // For MQTT actions
	Value    interface{} `json:"value"`
}

// Actions is a slice of Action that implements Scanner and Valuer for GORM
type Actions []Action

func (a Actions) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Actions) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

// Scene represents a collection of actions that can be triggered together
type Scene struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	TeraluxID string         `gorm:"type:char(36);not null;index" json:"teralux_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Actions   Actions        `gorm:"type:text;not null" json:"actions"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for the Scene model
func (Scene) TableName() string {
	return "scenes"
}
