package entities

import (
	"time"
)

type Report struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	S3ObjectKey  string    `json:"s3_object_key" gorm:"column:s3_object_key"`
	LocalPath    string    `json:"local_path"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Report) TableName() string {
	return "reports"
}
