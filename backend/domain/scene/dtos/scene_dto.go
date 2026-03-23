package dtos

// ActionDTO represents an action in request/response
type ActionDTO struct {
	DeviceID string      `json:"device_id,omitempty"`
	Code     string      `json:"code,omitempty"`
	RemoteID string      `json:"remote_id,omitempty"`
	Topic    string      `json:"topic,omitempty"`
	Value    interface{} `json:"value"`
}

// CreateSceneRequestDTO for POST /api/terminal/:id/scenes
type CreateSceneRequestDTO struct {
	Name    string      `json:"name" binding:"required"`
	Actions []ActionDTO `json:"actions"`
}

// UpdateSceneRequestDTO for PUT /api/terminal/:id/scenes/:scene_id
type UpdateSceneRequestDTO struct {
	Name    string      `json:"name" binding:"required" example:"Evening Mode"`
	Actions []ActionDTO `json:"actions"`
}

// SceneResponseDTO for GET /api/terminal/:id/scenes (includes terminal_id)
type SceneResponseDTO struct {
	ID         string      `json:"id"`
	TerminalID string      `json:"terminal_id"`
	Name       string      `json:"name"`
	Actions    []ActionDTO `json:"actions"`
}

// SceneListResponseDTO for summarized list
type SceneListResponseDTO struct {
	ID         string `json:"id"`
	TerminalID string `json:"terminal_id"`
	Name       string `json:"name"`
}

// SceneIDResponseDTO for returning just the scene ID
type SceneIDResponseDTO struct {
	SceneID string `json:"scene_id"`
}

// SceneItemDTO is a slim scene used inside TerminalScenesDTO (no terminal_id, it's implied by the wrapper)
type SceneItemDTO struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Actions []ActionDTO `json:"actions"`
}

// TerminalScenesDTO holds terminal_id and its scenes — used inside the wrapper
type TerminalScenesDTO struct {
	TerminalID string         `json:"terminal_id"`
	Scenes     []SceneItemDTO `json:"scenes"`
}

// TerminalScenesWrapperDTO wraps TerminalScenesDTO under "terminal" key
// Matches contract: { "terminal": { "terminal_id": "1", "scenes": [...] } }
type TerminalScenesWrapperDTO struct {
	Terminal TerminalScenesDTO `json:"terminal"`
}
