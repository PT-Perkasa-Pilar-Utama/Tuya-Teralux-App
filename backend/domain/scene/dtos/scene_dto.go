package dtos

// ActionDTO represents an action in request/response
type ActionDTO struct {
	DeviceID string      `json:"device_id,omitempty"`
	Code     string      `json:"code,omitempty"`
	RemoteID string      `json:"remote_id,omitempty"`
	Topic    string      `json:"topic,omitempty"`
	Value    interface{} `json:"value"`
}

// CreateSceneRequestDTO for POST /api/teralux/:id/scenes
type CreateSceneRequestDTO struct {
	Name    string      `json:"name" binding:"required"`
	Actions []ActionDTO `json:"actions"`
}

// UpdateSceneRequestDTO for PUT /api/teralux/:id/scenes/:scene_id
type UpdateSceneRequestDTO struct {
	Name    string      `json:"name" binding:"required"`
	Actions []ActionDTO `json:"actions"`
}

// SceneResponseDTO for GET /api/teralux/:id/scenes (includes teralux_id)
type SceneResponseDTO struct {
	ID        string      `json:"id"`
	TeraluxID string      `json:"teralux_id"`
	Name      string      `json:"name"`
	Actions   []ActionDTO `json:"actions"`
}

// SceneListResponseDTO for summarized list
type SceneListResponseDTO struct {
	ID        string `json:"id"`
	TeraluxID string `json:"teralux_id"`
	Name      string `json:"name"`
}

// SceneIDResponseDTO for returning just the scene ID
type SceneIDResponseDTO struct {
	SceneID string `json:"scene_id"`
}

// SceneItemDTO is a slim scene used inside TeraluxScenesDTO (no teralux_id, it's implied by the wrapper)
type SceneItemDTO struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Actions []ActionDTO `json:"actions"`
}

// TeraluxScenesDTO holds teralux_id and its scenes — used inside the wrapper
type TeraluxScenesDTO struct {
	TeraluxID string         `json:"teralux_id"`
	Scenes    []SceneItemDTO `json:"scenes"`
}

// TeraluxScenesWrapperDTO wraps TeraluxScenesDTO under "teralux" key
// Matches contract: { "teralux": { "teralux_id": "1", "scenes": [...] } }
type TeraluxScenesWrapperDTO struct {
	Teralux TeraluxScenesDTO `json:"teralux"`
}
