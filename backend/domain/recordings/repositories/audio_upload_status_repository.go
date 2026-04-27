package repositories

import (
	"gorm.io/gorm"

	"sensio/domain/infrastructure"
	"sensio/domain/recordings/entities"
)

type AudioUploadStatusRepository interface {
	Save(status *entities.AudioUploadStatus) error
	GetByObjectKey(objectKey string) (*entities.AudioUploadStatus, error)
	UpdateStatus(objectKey, status string) error
}

type audioUploadStatusRepository struct {
	db *gorm.DB
}

func NewAudioUploadStatusRepository() AudioUploadStatusRepository {
	return &audioUploadStatusRepository{
		db: infrastructure.DB,
	}
}

func (r *audioUploadStatusRepository) Save(status *entities.AudioUploadStatus) error {
	return r.db.Create(status).Error
}

func (r *audioUploadStatusRepository) GetByObjectKey(objectKey string) (*entities.AudioUploadStatus, error) {
	var status entities.AudioUploadStatus
	if err := r.db.Where("object_key = ?", objectKey).First(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *audioUploadStatusRepository) UpdateStatus(objectKey, status string) error {
	return r.db.Model(&entities.AudioUploadStatus{}).
		Where("object_key = ?", objectKey).
		Update("status", status).Error
}
