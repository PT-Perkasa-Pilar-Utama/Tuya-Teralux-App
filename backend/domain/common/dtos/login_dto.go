package dtos

type LoginRequestDTO struct {
	TerminalID string `json:"terminal_id" example:"550e8400-e29b-41d4-a716-446655440000" binding:"required"`
}

type LoginResponseDTO struct {
	TerminalID string `json:"terminal_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message    string `json:"message" example:"Login successful"`
}