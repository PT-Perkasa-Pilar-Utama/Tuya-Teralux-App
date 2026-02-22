package recordings_dtos

import (
	"time"
)

// RecordingResponseDto represents the recording data sent to the client
type RecordingResponseDto struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	AudioUrl     string    `json:"audio_url"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetAllRecordingsResponseDto represents the paginated response for getAll
type GetAllRecordingsResponseDto struct {
	Recordings []RecordingResponseDto `json:"recordings"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
}

type StandardResponse struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
	// Details is only populated for 400 (Bad Request) and 422 (Unprocessable Entity) errors.
	// For all other status codes, including 500, this field is nil/omitted.
	Details interface{} `json:"details,omitempty"`
}
