package dtos

// V1WhisperUploadSessionResponseDTO mirrors services.UploadSessionResponse for OpenAPI documentation
type V1WhisperUploadSessionResponseDTO struct {
	SessionID      string   `json:"session_id" example:"session_abc"`
	State          string   `json:"state" example:"uploading"` // uploading, ready, consumed, aborted, expired
	TotalChunks    int      `json:"total_chunks" example:"10"`
	ChunkSizeBytes int      `json:"chunk_size_bytes" example:"1048576"`
	TotalSizeBytes int64    `json:"total_size_bytes" example:"10485760"`
	ReceivedBytes  int64    `json:"received_bytes" example:"0"`
	MissingRanges  []string `json:"missing_ranges,omitempty" example:"[\"0-2\", \"5\"]"`
	ExpiresAt      int64    `json:"expires_at" example:"1741682064"`
}

// V1WhisperUploadChunkResponseDTO mirrors services.UploadChunkResponse for OpenAPI documentation
type V1WhisperUploadChunkResponseDTO struct {
	ReceivedChunks int    `json:"received_chunks" example:"1"`
	ReceivedBytes  int64  `json:"received_bytes" example:"1048576"`
	IsDuplicate    bool   `json:"is_duplicate" example:"false"`
	State          string `json:"state" example:"uploading"`
}

// CreateUploadSessionRequest used to initiate a new resumable upload
type CreateUploadSessionRequest struct {
	FileName       string `json:"file_name" binding:"required"`
	TotalSizeBytes int64  `json:"total_size_bytes" binding:"required"`
	ChunkSizeBytes int    `json:"chunk_size_bytes"`
	MimeType       string `json:"mime_type"`
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
	ExpiresAt      int64    `json:"expires_at"`
}

// UploadChunkAckDTO returned after a chunk is successfully received
type UploadChunkAckDTO struct {
	ReceivedChunks int    `json:"received_chunks"`
	ReceivedBytes  int64  `json:"received_bytes"`
	IsDuplicate    bool   `json:"is_duplicate"`
	State          string `json:"state"`
}
