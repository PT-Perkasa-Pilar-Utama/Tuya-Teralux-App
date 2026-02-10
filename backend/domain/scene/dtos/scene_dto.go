package dtos

// ActionDTO represents an action in request/response
type ActionDTO struct {
	DeviceID string      `json:"device_id,omitempty"`
	Code     string      `json:"code,omitempty"`
	RemoteID string      `json:"remote_id,omitempty"`
	Topic    string      `json:"topic,omitempty"`
	Value    interface{} `json:"value"`
}

// CreateSceneRequestDTO for POST /api/scenes
type CreateSceneRequestDTO struct {
	Name    string      `json:"name" binding:"required"`
	Actions []ActionDTO `json:"actions"`
}

// UpdateSceneRequestDTO for PUT /api/scenes/{id}
type UpdateSceneRequestDTO struct {
	Name    string      `json:"name" binding:"required"`
	Actions []ActionDTO `json:"actions"`
}

// SceneResponseDTO for response data
type SceneResponseDTO struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Actions []ActionDTO `json:"actions"`
}

// SceneListResponseDTO for summarized list
type SceneListResponseDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
