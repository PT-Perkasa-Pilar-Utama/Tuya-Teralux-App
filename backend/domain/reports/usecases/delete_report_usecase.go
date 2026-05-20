package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"sensio/domain/common/infrastructure"
	"sensio/domain/reports/repositories"
)

type DeleteReportUseCase interface {
	DeleteReport(id string) error
}

type deleteReportUseCase struct {
	repo       repositories.ReportRepository
	s3Service  *infrastructure.S3Service
}

func NewDeleteReportUseCase(repo repositories.ReportRepository, s3Service *infrastructure.S3Service) DeleteReportUseCase {
	return &deleteReportUseCase{repo: repo, s3Service: s3Service}
}

func (uc *deleteReportUseCase) DeleteReport(id string) error {
	report, err := uc.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("report not found: %v", err)
	}

	if report.S3ObjectKey != "" && uc.s3Service != nil {
		if err := uc.s3Service.DeleteObject(report.S3ObjectKey); err != nil {
			return fmt.Errorf("failed to delete S3 object: %v", err)
		}
	}

	if err := uc.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %v", err)
	}

	if report.LocalPath != "" {
		filePath := filepath.Join(".", report.LocalPath)
		_ = os.Remove(filePath)
	}

	return nil
}
