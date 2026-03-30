package dtos

// UpdateTerminalAIEngineProfileRequestDTO is the request body for updating a terminal's AI engine profile.
type UpdateTerminalAIEngineProfileRequestDTO struct {
	Profile *string `json:"profile"` // "fast", "standard", or null/empty to clear
}

// TerminalAIEngineProfileResponseDTO is the response shape for AI engine profile endpoints.
type TerminalAIEngineProfileResponseDTO struct {
	TerminalID        string  `json:"terminal_id"`
	Profile           *string `json:"profile"`
	Source            string  `json:"source"`
	EffectiveProvider *string `json:"effective_provider"`
	EffectiveMode     string  `json:"effective_mode"`
}
