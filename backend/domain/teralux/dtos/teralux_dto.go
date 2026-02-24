package dtos

import "time"

// CreateTeraluxRequestDTO represents the request body for creating a new teralux
type CreateTeraluxRequestDTO struct {
	MacAddress   string `json:"mac_address" binding:"required"`
	RoomID       string `json:"room_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	DeviceTypeID string `json:"device_type_id" binding:"required"`
}

// CreateTeraluxResponseDTO represents the response for creating a teralux (only returns ID)
type CreateTeraluxResponseDTO struct {
	TeraluxID string `json:"teralux_id"`
}

// UpdateTeraluxRequestDTO represents the request body for updating a teralux
type UpdateTeraluxRequestDTO struct {
	RoomID     *string `json:"room_id,omitempty"`
	MacAddress *string `json:"mac_address,omitempty"`
	Name       *string `json:"name,omitempty"`
}

// TeraluxFilterDTO represents filter options for listing teralux
type TeraluxFilterDTO struct {
	RoomID  *string `form:"room_id" json:"room_id,omitempty"`
	Page    int     `form:"page" json:"page,omitempty"`
	Limit   int     `form:"limit" json:"limit,omitempty"`
	PerPage int     `form:"per_page" json:"per_page,omitempty"` // Alias for limit
}

// TeraluxResponseDTO represents the response format for a single teralux
type TeraluxResponseDTO struct {
	ID         string    `json:"id"`
	MacAddress string    `json:"mac_address"`
	RoomID     string    `json:"room_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Devices:    []DeviceResponseDTO `json:"devices,omitempty"` // Removed to match test scenario
}

// TeraluxListResponseDTO represents the response format for a list of teralux items
type TeraluxListResponseDTO struct {
	Teralux []TeraluxResponseDTO `json:"teralux"`
	Total   int64                `json:"total"`
	Page    int                  `json:"page"`
	PerPage int                  `json:"per_page"`
}

// TeraluxSingleResponseDTO represents the response format for a single teralux (wrapped)
type TeraluxSingleResponseDTO struct {
	Teralux TeraluxResponseDTO `json:"teralux"`
}
