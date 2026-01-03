package dtos

// TuyaSyncDeviceDTO represents the simplified device status for sync responses
type TuyaSyncDeviceDTO struct {
	ID         string `json:"id"`
	RemoteID   string `json:"remote_id,omitempty"`
	Online     bool   `json:"online"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}
