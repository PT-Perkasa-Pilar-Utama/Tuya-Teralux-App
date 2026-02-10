package entities

// Action represents a single instruction within a scene
type Action struct {
	DeviceID string      `json:"device_id,omitempty"`
	Code     string      `json:"code,omitempty"`
	RemoteID string      `json:"remote_id,omitempty"` // For IR devices
	Topic    string      `json:"topic,omitempty"`     // For MQTT actions
	Value    interface{} `json:"value"`
}

// Scene represents a collection of actions that can be triggered together
type Scene struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}
