package usecases

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"sensio/domain/common/infrastructure"
	"sensio/domain/reports/entities"
	"sensio/domain/reports/repositories"
)

type SaveReportUseCase interface {
	SaveReportFromPath(srcPath, originalName string) (*entities.Report, error)
}

type saveReportUseCase struct {
	repo        repositories.ReportRepository
	fileService infrastructure.FileService
	s3Service   *infrastructure.S3Service
}

func NewSaveReportUseCase(repo repositories.ReportRepository, fileService infrastructure.FileService, s3Service *infrastructure.S3Service) SaveReportUseCase {
	return &saveReportUseCase{
		repo:        repo,
		fileService: fileService,
		s3Service:   s3Service,
	}
}

func (uc *saveReportUseCase) SaveReportFromPath(srcPath, originalName string) (*entities.Report, error) {
	fileExt := filepath.Ext(originalName)
	if fileExt == "" {
		fileExt = ".pdf"
	}
	uuidFilename, _ := uuid.NewV7()
	newFilename := uuidFilename.String() + fileExt

	var s3ObjectKey string
	var localPath string

	if uc.s3Service != nil {
		objectKey := uc.s3Service.BuildObjectKey("reports", newFilename)
		contentType := mime.TypeByExtension(fileExt)
		if contentType == "" {
			contentType = "application/pdf"
		}
		if _, err := uc.s3Service.UploadFile(srcPath, objectKey, contentType); err != nil {
			return nil, fmt.Errorf("failed to upload to S3: %w", err)
		}
		s3ObjectKey = objectKey
		_ = os.Remove(srcPath)
	} else {
		uploadPath := filepath.Join("uploads", "reports", newFilename)
		if err := uc.fileService.MoveFile(srcPath, uploadPath); err != nil {
			return nil, fmt.Errorf("failed to move file: %v", err)
		}
		localPath = uploadPath
	}

	uuidEntity, _ := uuid.NewV7()
	report := &entities.Report{
		ID:           uuidEntity.String(),
		Filename:     newFilename,
		OriginalName: originalName,
		S3ObjectKey:  s3ObjectKey,
		LocalPath:    localPath,
		CreatedAt:    time.Now(),
	}

	if err := uc.repo.Save(report); err != nil {
		return nil, err
	}

	return report, nil
}
