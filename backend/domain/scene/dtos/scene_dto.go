package dtos

// ActionDTO represents an action in request/response
type ActionDTO struct {
	DeviceID string      `json:"device_id" binding:"required,uuid"`
	Code     string      `json:"code" binding:"required,max=100"`
	RemoteID string      `json:"remote_id" binding:"omitempty,max=255"`
	Topic    string      `json:"topic" binding:"omitempty,max=255"`
	Value    interface{} `json:"value"`
}

// CreateSceneRequestDTO for POST /api/terminal/:id/scenes
type CreateSceneRequestDTO struct {
	Name        string      `json:"name" binding:"required,min=1,max=255"`
	Description string      `json:"description" binding:"omitempty,max=1000"`
	Actions     []ActionDTO `json:"actions" binding:"required,min=1,dive"`
}

// UpdateSceneRequestDTO for PUT /api/terminal/:id/scenes/:scene_id
type UpdateSceneRequestDTO struct {
	Name        string      `json:"name" binding:"required" example:"Evening Mode"`
	Description string      `json:"description" binding:"omitempty,max=1000"`
	Actions     []ActionDTO `json:"actions"`
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
