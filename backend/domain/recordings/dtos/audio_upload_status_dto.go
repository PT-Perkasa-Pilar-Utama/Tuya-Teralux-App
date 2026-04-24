package recordings_dtos

type AudioUploadStatusDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ObjectKey string `json:"object_key"`
	Status    string `json:"status"`
}

type UpdateAudioUploadStatusRequest struct {
	ObjectKey string `json:"object_key" binding:"required"`
	Status    string `json:"status" binding:"required,oneof=PENDING COMPLETED FAILED"`
}
