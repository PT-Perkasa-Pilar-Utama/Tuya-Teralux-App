package usecases

import (
	"sensio/domain/common/infrastructure"
	reports_dtos "sensio/domain/reports/dtos"
	"sensio/domain/reports/repositories"
)

type GetReportByIDUseCase interface {
	GetReportByID(id string) (*reports_dtos.ReportResponseDto, error)
}

type getReportByIDUseCase struct {
	repo      repositories.ReportRepository
	s3Service *infrastructure.S3Service
}

func NewGetReportByIDUseCase(repo repositories.ReportRepository, s3Service *infrastructure.S3Service) GetReportByIDUseCase {
	return &getReportByIDUseCase{repo: repo, s3Service: s3Service}
}

func (uc *getReportByIDUseCase) GetReportByID(id string) (*reports_dtos.ReportResponseDto, error) {
	report, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &reports_dtos.ReportResponseDto{
		ID:           report.ID,
		Filename:     report.Filename,
		OriginalName: report.OriginalName,
		CreatedAt:    report.CreatedAt,
	}, nil
}
