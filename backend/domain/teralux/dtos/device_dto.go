package dtos

import "time"

// CreateDeviceRequestDTO represents the request body for creating a new device
type CreateDeviceRequestDTO struct {
	ID        string `json:"id" binding:"required"`
	TeraluxID string `json:"teralux_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// CreateDeviceResponseDTO represents the response for creating a device (only returns ID)
type CreateDeviceResponseDTO struct {
	DeviceID string `json:"device_id"`
}

// UpdateDeviceRequestDTO represents the request body for updating a device
type UpdateDeviceRequestDTO struct {
	Name *string `json:"name,omitempty"`
}

// DeviceFilterDTO represents filter options for listing devices
type DeviceFilterDTO struct {
	TeraluxID *string `form:"teralux_id" json:"teralux_id,omitempty"`
	Page      int     `form:"page" json:"page,omitempty"`
	Limit     int     `form:"limit" json:"limit,omitempty"`
	PerPage   int     `form:"per_page" json:"per_page,omitempty"` // Alias for limit
}

// DeviceResponseDTO represents the response format for a single device
type DeviceResponseDTO struct {
	ID                string    `json:"id"`
	TeraluxID         string    `json:"teralux_id"`
	Name              string    `json:"name"`
	RemoteID          string    `json:"remote_id"`
	Category          string    `json:"category"`
	RemoteCategory    string    `json:"remote_category"`
	ProductName       string    `json:"product_name"`
	RemoteProductName string    `json:"remote_product_name"`
	Icon              string    `json:"icon"`
	CustomName        string    `json:"custom_name"`
	Model             string    `json:"model"`
	IP                string    `json:"ip"`
	LocalKey          string    `json:"local_key"`
	GatewayID         string    `json:"gateway_id"`
	CreateTime        int64     `json:"create_time"`
	UpdateTime        int64     `json:"update_time"`
	Collections       string    `json:"collections"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// DeviceStatusDTO represents a device status for responses
type DeviceStatusDTO struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

// DeviceListResponseDTO represents the response format for a list of devices
type DeviceListResponseDTO struct {
	Devices []DeviceResponseDTO `json:"devices"`
	Total   int                 `json:"total"`
	Page    int                 `json:"page"`
	PerPage int                 `json:"per_page"`
}

// DeviceSingleResponseDTO represents the response format for a single device (wrapped)
type DeviceSingleResponseDTO struct {
	Device DeviceResponseDTO `json:"device"`
}
