package usecases

import (
	"context"
	"fmt"

	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/entities"
	"sensio/domain/recordings/repositories"
)

type UpdateAudioUploadStatusUseCase interface {
	UpdateStatus(ctx context.Context, req recordings_dtos.UpdateAudioUploadStatusRequest) error
}

type updateAudioUploadStatusUseCase struct {
	repo repositories.AudioUploadStatusRepository
}

func NewUpdateAudioUploadStatusUseCase(repo repositories.AudioUploadStatusRepository) UpdateAudioUploadStatusUseCase {
	return &updateAudioUploadStatusUseCase{
		repo: repo,
	}
}

func (u *updateAudioUploadStatusUseCase) UpdateStatus(ctx context.Context, req recordings_dtos.UpdateAudioUploadStatusRequest) error {
	if req.ObjectKey == "" {
		return fmt.Errorf("object_key is required")
	}
	if req.Status != entities.AudioUploadStatusPending &&
		req.Status != entities.AudioUploadStatusCompleted &&
		req.Status != entities.AudioUploadStatusFailed {
		return fmt.Errorf("invalid status: must be PENDING, COMPLETED, or FAILED")
	}

	return u.repo.UpdateStatus(req.ObjectKey, req.Status)
}
