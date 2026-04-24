package recordings_dtos

// CreateUploadIntentRequest used to request a signed URL for direct S3 upload
type CreateUploadIntentRequest struct {
	Filename    string `json:"filename" binding:"required"`
	Size        int64  `json:"size" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	BookingID   string `json:"booking_id" binding:"required"`
}

// UploadIntentResponseDTO returned with presigned S3 URL for direct upload
type UploadIntentResponseDTO struct {
	ObjectKey   string `json:"object_key"`
	UploadURL   string `json:"presigned_url"`
	ContentType string `json:"content_type"`
	ExpiresAt   string `json:"expires_at"`
}
