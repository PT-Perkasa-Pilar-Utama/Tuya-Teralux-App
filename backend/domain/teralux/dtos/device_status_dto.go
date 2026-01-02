package dtos

import "time"

// CreateDeviceStatusRequestDTO represents the request body for creating a new device status
type CreateDeviceStatusRequestDTO struct {
	DeviceID string `json:"device_id" binding:"required"`
	Name     string `json:"name,omitempty"`
	Code     string `json:"code" binding:"required"`
	Value    int    `json:"value"`
}

// CreateDeviceStatusResponseDTO represents the response for creating a device status
type CreateDeviceStatusResponseDTO struct {
	ID string `json:"status_id"`
}

// UpdateDeviceStatusRequestDTO represents the request body for updating a device status
type UpdateDeviceStatusRequestDTO struct {
	Name  string `json:"name,omitempty"`
	Value *int   `json:"value,omitempty"` // Pointer to allow 0 as a valid value update if needed, though int is fine if 0 is not 'empty'
}

// DeviceStatusResponseDTO represents the response format for a single device status
type DeviceStatusResponseDTO struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Value     int       `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceStatusListResponseDTO represents the response format for a list of device statuses
type DeviceStatusListResponseDTO struct {
	Statuses []DeviceStatusResponseDTO `json:"statuses"`
	Total    int                       `json:"total"`
}
