package dtos

// CreateUploadSessionRequest used to initiate a new resumable upload
type CreateUploadSessionRequest struct {
	FileName       string `json:"file_name" binding:"required"`
	TotalSizeBytes int64  `json:"total_size_bytes" binding:"required"`
	ChunkSizeBytes int    `json:"chunk_size_bytes"`
	MimeType       string `json:"mime_type"`
	OwnerUID       string `json:"-"` // Set from auth context
}

// UploadSessionResponseDTO returned after session creation or status check
type UploadSessionResponseDTO struct {
	SessionID      string   `json:"session_id"`
	State          string   `json:"state"` // uploading, ready, consumed, aborted, expired
	TotalChunks    int      `json:"total_chunks"`
	ChunkSizeBytes int      `json:"chunk_size_bytes"`
	TotalSizeBytes int64    `json:"total_size_bytes"`
	ReceivedBytes  int64    `json:"received_bytes"`
	MissingRanges  []string `json:"missing_ranges,omitempty"` // e.g. ["0-2", "5"]
	ExpiresAt      string   `json:"expires_at"`
}

// UploadChunkAckDTO returned after a chunk is successfully received
type UploadChunkAckDTO struct {
	ReceivedChunks int    `json:"received_chunks"`
	ReceivedBytes  int64  `json:"received_bytes"`
	IsDuplicate    bool   `json:"is_duplicate"`
	State          string `json:"state"`
}

// SubmitByUploadRequest used to start a transcribe job from a completed upload session
type SubmitByUploadRequest struct {
	SessionID      string `json:"session_id" binding:"required"`
	Language       string `json:"language"`
	MacAddress     string `json:"mac_address"`
	Diarize        bool   `json:"diarize"`
	IdempotencyKey string `json:"idempotency_key"`
}

// PipelineSubmitByUploadRequest used to start a pipeline job from a completed upload session
type PipelineSubmitByUploadRequest struct {
	SessionID      string   `json:"session_id" binding:"required"`
	Language       string   `json:"language"`
	TargetLanguage string   `json:"target_language"`
	MacAddress     string   `json:"mac_address"`
	Diarize        bool     `json:"diarize"`
	Refine         *bool    `json:"refine"`
	Summarize      bool     `json:"summarize"`
	Participants   []string `json:"participants"`
	Context        string   `json:"context"`
	Style          string   `json:"style"`
	Date           string   `json:"date"`
	Location       string   `json:"location"`
	IdempotencyKey string   `json:"idempotency_key"`
}
