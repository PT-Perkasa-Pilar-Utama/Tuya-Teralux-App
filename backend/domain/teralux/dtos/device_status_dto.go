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
	DeviceID string `json:"device_id"`
	Code     string `json:"code"`
}

// UpdateDeviceStatusRequestDTO represents the request body for updating a device status
type UpdateDeviceStatusRequestDTO struct {
	Code  string      `json:"code" binding:"required"`
	Value interface{} `json:"value,omitempty"`
}

// DeviceStatusResponseDTO represents the response format for a single device status
type DeviceStatusResponseDTO struct {
	DeviceID  string    `json:"device_id"`
	Code      string    `json:"code"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceStatusListResponseDTO represents the response format for a list of device statuses
type DeviceStatusListResponseDTO struct {
	DeviceStatuses []DeviceStatusResponseDTO `json:"device_statuses"`
	Total          int                       `json:"total"`
	Page           int                       `json:"page"`
	PerPage        int                       `json:"per_page"`
}

// DeviceStatusSingleResponseDTO represents the response format for a single device status (wrapped)
type DeviceStatusSingleResponseDTO struct {
	DeviceStatus DeviceStatusResponseDTO `json:"device_status"`
}
