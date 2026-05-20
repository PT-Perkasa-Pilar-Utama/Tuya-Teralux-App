package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"sensio/domain/common/infrastructure"
	"sensio/domain/recordings/repositories"
)

type DeleteRecordingUseCase interface {
	DeleteRecording(id string) error
}

type deleteRecordingUseCase struct {
	repo       repositories.RecordingRepository
	s3Service  *infrastructure.S3Service
}

func NewDeleteRecordingUseCase(repo repositories.RecordingRepository, s3Service *infrastructure.S3Service) DeleteRecordingUseCase {
	return &deleteRecordingUseCase{repo: repo, s3Service: s3Service}
}

func (uc *deleteRecordingUseCase) DeleteRecording(id string) error {
	recording, err := uc.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("recording not found: %v", err)
	}

	if recording.S3ObjectKey != "" && uc.s3Service != nil {
		if err := uc.s3Service.DeleteObject(recording.S3ObjectKey); err != nil {
			return fmt.Errorf("failed to delete S3 object: %v", err)
		}
	}

	if err := uc.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %v", err)
	}

	if recording.Filename != "" {
		filePath := filepath.Join("uploads", "audio", filepath.Base(recording.Filename))
		_ = os.Remove(filePath)
	}

	return nil
}
