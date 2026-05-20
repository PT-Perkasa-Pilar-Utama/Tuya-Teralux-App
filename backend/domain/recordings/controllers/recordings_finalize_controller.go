package controllers

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/entities"
	"sensio/domain/recordings/repositories"
	"sensio/domain/recordings/services"
)

type RecordingsFinalizeController struct {
	repo            repositories.RecordingRepository
	s3Service       *infrastructure.S3Service
	bigAudioService services.BIGRoomAudioUpdateService
}

func NewRecordingsFinalizeController(
	repo repositories.RecordingRepository,
	s3Service *infrastructure.S3Service,
	bigAudioService services.BIGRoomAudioUpdateService,
) *RecordingsFinalizeController {
	return &RecordingsFinalizeController{
		repo:            repo,
		s3Service:       s3Service,
		bigAudioService: bigAudioService,
	}
}

type FinalizeRequest struct {
	Filename   string `json:"filename" binding:"required"`
	ObjectKey  string `json:"object_key" binding:"required"`
	MacAddress string `json:"mac_address" binding:"required"`
}

// FinalizeRecording handles POST /api/recordings/finalize
// @Summary Finalize an S3-uploaded recording
// @Description Persists recording metadata after S3 upload completes and generates presigned GET URL for playback
// @Tags 06. Recordings
// @Accept json
// @Produce json
// @Param body body FinalizeRequest true "Finalize recording request"
// @Success 200 {object} commonDtos.StandardResponse{data=recordings_dtos.RecordingResponseDto}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse
// @Failure 503 {object} commonDtos.StandardResponse
// @Router /api/recordings/finalize [post]
func (c *RecordingsFinalizeController) FinalizeRecording(ctx *gin.Context) {
	if c.s3Service == nil {
		utils.LogError("RecordingsFinalizeController.FinalizeRecording: S3 service unavailable")
		ctx.JSON(http.StatusServiceUnavailable, commonDtos.StandardResponse{
			Status:  false,
			Message: "S3 service unavailable",
		})
		return
	}

	var req FinalizeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "invalid request body",
		})
		return
	}

	expectedPrefix := strings.TrimSuffix(c.s3Service.BuildObjectKey("recordings", "placeholder"), "placeholder")
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "invalid object key",
		})
		return
	}

	presignedGETURL, err := c.s3Service.GeneratePresignedGetURL(req.ObjectKey)
	if err != nil {
		utils.LogError("RecordingsFinalizeController.FinalizeRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "failed to generate presigned URL",
		})
		return
	}

	storedFilename := filepath.Base(req.ObjectKey)

	recording := entities.Recording{
		ID:           uuid.New().String(),
		Filename:     storedFilename,
		OriginalName: req.Filename,
		AudioUrl:     presignedGETURL,
		S3ObjectKey:  req.ObjectKey,
		MacAddress:   req.MacAddress,
		CreatedAt:    time.Now(),
	}

	if err := c.repo.Save(&recording); err != nil {
		utils.LogError("RecordingsFinalizeController.FinalizeRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "failed to save recording",
		})
		return
	}

	if recording.MacAddress != "" {
		go func() {
			if err := c.bigAudioService.UpdateRoomOccupiedAudio(recording.MacAddress, recording.AudioUrl); err != nil {
				utils.LogError("RecordingsFinalizeController.FinalizeRecording: BIG audio update failed: %v", err)
			}
		}()
	}

	response := recordings_dtos.RecordingResponseDto{
		ID:           recording.ID,
		Filename:     recording.Filename,
		OriginalName: recording.OriginalName,
		AudioUrl:     recording.AudioUrl,
		CreatedAt:    recording.CreatedAt,
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Recording finalized successfully",
		Data:    response,
	})
}