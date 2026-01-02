package dtos

import "time"

// CreateDeviceRequestDTO represents the request body for creating a new device
type CreateDeviceRequestDTO struct {
	TeraluxID string `json:"teralux_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// CreateDeviceResponseDTO represents the response for creating a device
type CreateDeviceResponseDTO struct {
	ID string `json:"device_id"`
}

// UpdateDeviceRequestDTO represents the request body for updating a device
type UpdateDeviceRequestDTO struct {
	Name string `json:"name,omitempty"`
}

// DeviceResponseDTO represents the response format for a single device
type DeviceResponseDTO struct {
	ID        string    `json:"id"`
	TeraluxID string    `json:"teralux_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceListResponseDTO represents the response format for a list of devices
type DeviceListResponseDTO struct {
	Devices []DeviceResponseDTO `json:"devices"`
	Total   int                 `json:"total"`
}
