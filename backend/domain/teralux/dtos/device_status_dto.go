package dtos

import "time"

// CreateDeviceStatusRequestDTO represents the request body for creating a new device status
type CreateDeviceStatusRequestDTO struct {
	DeviceID string `json:"device_id" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Value    string `json:"value"`
}

// CreateDeviceStatusResponseDTO represents the response for creating a device status
type CreateDeviceStatusResponseDTO struct {
	ID string `json:"status_id"`
}

// UpdateDeviceStatusRequestDTO represents the request body for updating a device status
type UpdateDeviceStatusRequestDTO struct {
	Value string `json:"value,omitempty"`
}

// DeviceStatusResponseDTO represents the response format for a single device status
type DeviceStatusResponseDTO struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Code      string    `json:"code"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceStatusListResponseDTO represents the response format for a list of device statuses
type DeviceStatusListResponseDTO struct {
	Statuses []DeviceStatusResponseDTO `json:"statuses"`
	Total    int                       `json:"total"`
}
