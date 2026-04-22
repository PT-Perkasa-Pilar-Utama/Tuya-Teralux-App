package recordings_dtos



// RecordingResponseDto represents the recording data sent to the client
type RecordingResponseDto struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	AudioURL     string    `json:"audio_url"`
	CreatedAt    string    `json:"created_at"`
}

// GetAllRecordingsResponseDto represents the paginated response for getAll
type GetAllRecordingsResponseDto struct {
	Recordings []RecordingResponseDto `json:"recordings"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
}
