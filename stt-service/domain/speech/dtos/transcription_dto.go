package dtos

// StandardResponse represents the standardized API response structure, matching the main backend
type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Details interface{} `json:"details,omitempty"`
}

// TranscriptionResponseDTO represents the data returned after successful transcription
type TranscriptionResponseDTO struct {
	Text string `json:"text"`
}
