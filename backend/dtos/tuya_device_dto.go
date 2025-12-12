package dtos

// TuyaDeviceDTO represents a single device for API consumers
type TuyaDeviceDTO struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Category    string                `json:"category"`
	ProductName string                `json:"product_name"`
	Online      bool                  `json:"online"`
	Icon        string                `json:"icon"`
	Status      []TuyaDeviceStatusDTO `json:"status"`
	CustomName  string                `json:"custom_name,omitempty"`
	Model       string                `json:"model,omitempty"`
	IP          string                `json:"ip,omitempty"`
	LocalKey    string                `json:"local_key"`
	GatewayID   string                `json:"gateway_id"`
	CreateTime  int64                 `json:"create_time"`
	UpdateTime  int64                 `json:"update_time"`
}

// TuyaCommandDTO represents a single command
type TuyaCommandDTO struct {
	Code  string      `json:"code" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// TuyaCommandsRequestDTO represents the request body for sending commands
type TuyaCommandsRequestDTO struct {
	Commands []TuyaCommandDTO `json:"commands" binding:"required"`
}

// TuyaIRACCommandDTO represents a single IR AC command request
type TuyaIRACCommandDTO struct {
	InfraredID string `json:"infrared_id" binding:"required"`
	RemoteID   string `json:"remote_id" binding:"required"`
	Code       string `json:"code" binding:"required"`
	Value      int    `json:"value"`
}

// TuyaDeviceStatusDTO represents device status for API consumers
type TuyaDeviceStatusDTO struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// TuyaDevicesResponseDTO represents the response for getting all devices
type TuyaDevicesResponseDTO struct {
	Devices []TuyaDeviceDTO `json:"devices"`
	Total   int             `json:"total"`
}

// TuyaDeviceResponseDTO represents the response for getting a single device
type TuyaDeviceResponseDTO struct {
	Device TuyaDeviceDTO `json:"device"`
}
