package reports_dtos

import (
	"time"
)

type ReportResponseDto struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type UploadURLResponseDto struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	Filename  string `json:"filename"`
}
