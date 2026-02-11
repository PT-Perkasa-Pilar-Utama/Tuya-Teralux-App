package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"teralux_app/domain/recordings/repositories"
)

type DeleteRecordingUseCase interface {
	Execute(id string) error
}

type deleteRecordingUseCase struct {
	repo repositories.RecordingRepository
}

func NewDeleteRecordingUseCase(repo repositories.RecordingRepository) DeleteRecordingUseCase {
	return &deleteRecordingUseCase{repo: repo}
}

func (uc *deleteRecordingUseCase) Execute(id string) error {
	// 1. Get recording to find filename
	recording, err := uc.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("recording not found: %v", err)
	}

	// 2. Delete metadata from DB
	if err := uc.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %v", err)
	}

	// 3. Delete physical file (Hard Delete)
	// Filename is stored as UUID.EXT
	filePath := filepath.Join("uploads", "audio", recording.Filename)
	_ = os.Remove(filePath) // Ignore error if file doesn't exist, metadata is gone anyway

	return nil
}
