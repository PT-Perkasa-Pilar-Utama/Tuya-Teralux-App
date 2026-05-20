package controllers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"sensio/domain/common/dtos"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/reports/repositories"
)

var _ = dtos.StandardResponse{}

type ReportDownloadController struct {
	repo      repositories.ReportRepository
	s3Service *infrastructure.S3Service
}

func NewReportDownloadController(repo repositories.ReportRepository, s3Service *infrastructure.S3Service) *ReportDownloadController {
	return &ReportDownloadController{
		repo:      repo,
		s3Service: s3Service,
	}
}

func (c *ReportDownloadController) GetDownload(ctx *gin.Context) {
	id := ctx.Param("id")

	report, err := c.repo.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Report not found",
		})
		return
	}

	if report.S3ObjectKey != "" && c.s3Service != nil {
		presignedURL, err := c.s3Service.GeneratePresignedGetURL(report.S3ObjectKey)
		if err != nil {
			utils.LogError("ReportDownloadController.GetDownload: Failed to generate presigned URL: %v", err)
			ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Internal Server Error",
			})
			return
		}
		if presignedURL != "" {
			ctx.Redirect(http.StatusFound, presignedURL)
			return
		}
	}

	if report.LocalPath != "" {
		filePath := filepath.Join(".", report.LocalPath)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
				Status:  false,
				Message: "Report file not found",
			})
			return
		}
		ctx.File(filePath)
		return
	}

	ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
		Status:  false,
		Message: "Report file not available",
	})
}
