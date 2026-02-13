package usecases

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/recordings/entities"
	"teralux_app/domain/recordings/repositories"
)

type SaveRecordingUseCase interface {
	Execute(file *multipart.FileHeader) (*entities.Recording, error)
}

type saveRecordingUseCase struct {
	repo        repositories.RecordingRepository
	fileService infrastructure.FileService
}

func NewSaveRecordingUseCase(repo repositories.RecordingRepository, fileService infrastructure.FileService) SaveRecordingUseCase {
	return &saveRecordingUseCase{
		repo:        repo,
		fileService: fileService,
	}
}

func (uc *saveRecordingUseCase) Execute(fileHeader *multipart.FileHeader) (*entities.Recording, error) {
	// 1. Generate UUIDv4 for filename
	fileExt := filepath.Ext(fileHeader.Filename)
	newFilename := uuid.New().String() + fileExt
	
	// 2. Define paths
	uploadPath := filepath.Join("uploads", "audio", newFilename)
	
	// 3. Save physical file (using gin context indirectly or standard io)
	// Since we are in Usecase, we need to handle file saving responsibly.
	// In a clean architecture, file saving might be in infrastructure service, 
	// but for simplicity here we assume the controller passes the file handler or we use a helper.
	// HOWEVER, the standard way in Gin is via Context.SaveUploadedFile.
	// To keep Usecase pure, we should receive the *multipart.FileHeader and save it manually.
	
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	if err := uc.fileService.SaveUploadedFile(fileHeader, uploadPath); err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// 4. Construct Public URL
	// Assuming the server domain/port is handled by frontend proxy or configuration
	// But as per requirement: /uploads/audio/{filename}
	publicUrl := fmt.Sprintf("/uploads/audio/%s", newFilename)

	// 5. Create Entity
	recording := &entities.Recording{
		ID:           uuid.New().String(),
		Filename:     newFilename,
		OriginalName: fileHeader.Filename,
		AudioUrl:     publicUrl,
		CreatedAt:    time.Now(),
	}

	// 6. Save Metadata
	if err := uc.repo.Save(recording); err != nil {
		return nil, err
	}

	return recording, nil
}
