package dtos

import "time"

// CreateTerminalRequestDTO represents the request body for creating a new terminal
type CreateTerminalRequestDTO struct {
	MacAddress   string `json:"mac_address" binding:"required"`
	RoomID       string `json:"room_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	DeviceTypeID string `json:"device_type_id" binding:"required"`
}

// CreateTerminalResponseDTO represents the response for creating a terminal
type CreateTerminalResponseDTO struct {
	TerminalID   string `json:"terminal_id"`
	MQTTUsername string `json:"mqtt_username"`
	MQTTPassword string `json:"mqtt_password"`
}

// UpdateTerminalRequestDTO represents the request body for updating a terminal
type UpdateTerminalRequestDTO struct {
	RoomID       *string `json:"room_id,omitempty"`
	MacAddress   *string `json:"mac_address,omitempty"`
	Name         *string `json:"name,omitempty"`
	DeviceTypeID *string `json:"device_type_id,omitempty"`
}

// TerminalFilterDTO represents filter options for listing terminal
type TerminalFilterDTO struct {
	RoomID  *string `form:"room_id" json:"room_id,omitempty"`
	Page    int     `form:"page" json:"page,omitempty"`
	Limit   int     `form:"limit" json:"limit,omitempty"`
	PerPage int     `form:"per_page" json:"per_page,omitempty"` // Alias for limit
}

// TerminalResponseDTO represents the response format for a single terminal
type TerminalResponseDTO struct {
	ID           string    `json:"id"`
	MacAddress   string    `json:"mac_address"`
	RoomID       string    `json:"room_id"`
	Name         string    `json:"name"`
	DeviceTypeID string    `json:"device_type_id"`
	MQTTUsername string    `json:"mqtt_username"`
	MQTTPassword string    `json:"mqtt_password,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// Devices:    []DeviceResponseDTO `json:"devices,omitempty"` // Removed to match test scenario
}

// TerminalListResponseDTO represents the response format for a list of terminal items
type TerminalListResponseDTO struct {
	Terminal []TerminalResponseDTO `json:"terminal"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PerPage  int                   `json:"per_page"`
}

// TerminalSingleResponseDTO represents the response format for a single terminal (wrapped)
type TerminalSingleResponseDTO struct {
	Terminal TerminalResponseDTO `json:"terminal"`
}

// MQTTCredentialsResponseDTO represents MQTT credentials for device authentication
type MQTTCredentialsResponseDTO struct {
	Username string `json:"username" example:"device_001"`
	Password string `json:"password" example:"secret_password"`
}
