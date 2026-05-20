package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	recordings_dtos "sensio/domain/recordings/dtos"
)

type RecordingsPresignController struct {
	s3Service *infrastructure.S3Service
}

func NewRecordingsPresignController(s3Service *infrastructure.S3Service) *RecordingsPresignController {
	return &RecordingsPresignController{s3Service: s3Service}
}

type UploadURLRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

// CreateUploadURL handles POST /api/recordings/upload-url
// @Summary Get a presigned URL for uploading a recording
// @Description Generates a presigned S3 PUT URL for direct upload
// @Tags 06. Recordings
// @Accept json
// @Produce json
// @Param body body UploadURLRequest true "Upload URL request"
// @Success 200 {object} commonDtos.StandardResponse{data=recordings_dtos.UploadURLResponseDto}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 503 {object} commonDtos.StandardResponse
// @Router /api/recordings/upload-url [post]
func (c *RecordingsPresignController) CreateUploadURL(ctx *gin.Context) {
	if c.s3Service == nil {
		utils.LogError("RecordingsPresignController.CreateUploadURL: S3 service unavailable")
		ctx.JSON(http.StatusServiceUnavailable, commonDtos.StandardResponse{
			Status:  false,
			Message: "S3 service unavailable",
		})
		return
	}

	var req UploadURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "invalid request body",
		})
		return
	}

	ext := ""
	if idx := strings.LastIndex(req.Filename, "."); idx != -1 {
		ext = req.Filename[idx:]
	}

	newFilename := uuid.New().String() + ext
	objectKey := c.s3Service.BuildObjectKey("recordings", newFilename)

	uploadURL, err := c.s3Service.GeneratePresignedPutURL(objectKey, req.ContentType)
	if err != nil {
		utils.LogError("RecordingsPresignController.CreateUploadURL: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "failed to generate upload URL",
		})
		return
	}

	response := recordings_dtos.UploadURLResponseDto{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		Filename:  newFilename,
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Upload URL generated successfully",
		Data:    response,
	})
}