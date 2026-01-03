package dtos

import "time"

// CreateTeraluxRequestDTO represents the request body for creating a new teralux
type CreateTeraluxRequestDTO struct {
	MacAddress string `json:"mac_address" binding:"required"`
	RoomID     string `json:"room_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

// CreateTeraluxResponseDTO represents the response for creating a teralux (only returns ID)
type CreateTeraluxResponseDTO struct {
	ID string `json:"teralux_id"`
}

// UpdateTeraluxRequestDTO represents the request body for updating a teralux
type UpdateTeraluxRequestDTO struct {
	MacAddress string `json:"mac_address,omitempty"`
	Name       string `json:"name,omitempty"`
}

// TeraluxResponseDTO represents the response format for a single teralux
type TeraluxResponseDTO struct {
	ID         string    `json:"id"`
	MacAddress string    `json:"mac_address"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TeraluxListResponseDTO represents the response format for a list of teralux items
type TeraluxListResponseDTO struct {
	Teralux []TeraluxResponseDTO `json:"teralux"`
	Total   int                  `json:"total"`
}
