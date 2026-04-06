package dtos

// AssistantSafeDeviceDTO represents a single device for assistant/vector storage.
// Excludes sensitive fields: local_key, ip, gateway_id, icon, create_time, update_time
// NOTE: Online status is intentionally excluded to prevent LLM from making decisions based on stale cache data.
// Device status is fetched real-time during control execution.
type AssistantSafeDeviceDTO struct {
	ID             string                `json:"id"`
	RemoteID       string                `json:"remote_id,omitempty"`
	Name           string                `json:"name"`
	Category       string                `json:"category"`
	RemoteCategory string                `json:"remote_category,omitempty"`
	ProductName    string                `json:"product_name,omitempty"`
	Status         []TuyaDeviceStatusDTO `json:"status,omitempty"`
}

// AssistantSafeDevicesSnapshotDTO represents an assistant-safe snapshot of user's devices.
// This is stored in vector DB key: tuya:devices:uid:{uid}
// Only written for full (non-paginated, non-filtered) device list requests.
type AssistantSafeDevicesSnapshotDTO struct {
	Devices      []AssistantSafeDeviceDTO `json:"devices"`
	TotalDevices int                      `json:"total_devices"`
	UpdatedAt    int64                    `json:"updated_at"`
	Source       string                   `json:"source"` // "tuya_api_full_snapshot"
}
