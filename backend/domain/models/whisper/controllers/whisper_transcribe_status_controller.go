package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/tasks"
	"sensio/domain/models/whisper/dtos"

	"github.com/gin-gonic/gin"
)

// WhisperTranscribeStatusController handles GET /api/models/whisper/transcribe/:transcribe_id.
type WhisperTranscribeStatusController struct {
	statusUC tasks.GenericStatusUseCase[dtos.AsyncTranscriptionStatusDTO]
}

func NewWhisperTranscribeStatusController(statusUC tasks.GenericStatusUseCase[dtos.AsyncTranscriptionStatusDTO]) *WhisperTranscribeStatusController {
	return &WhisperTranscribeStatusController{
		statusUC: statusUC,
	}
}

// GetStatus handles GET /api/models/whisper/transcribe/:transcribe_id
// @Summary Get transcription status (Consolidated)
// @Description Get the status and result of any transcription task (Short, Long, or Orion).
// @Tags 04. Models
// @Security BearerAuth
// @Produce json
// @Param transcribe_id path string true "Task ID"
// @Success 200 {object} commonDtos.StandardResponse{data=dtos.AsyncTranscriptionStatusDTO}
// @Failure 404 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/models/whisper/transcribe/{transcribe_id} [get]
func (c *WhisperTranscribeStatusController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("transcribe_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Task ID is required",
		})
		return
	}

	status, err := c.statusUC.GetTaskStatus(taskID)
	if err == nil {
		isSuccess := status.Status != "failed"
		message := "Task status retrieved successfully"
		httpStatus := http.StatusOK

		if status.Status == "failed" {
			message = "Task failed: " + status.Error
			if status.HTTPStatusCode != 0 {
				httpStatus = status.HTTPStatusCode
			}
		}

		ctx.JSON(httpStatus, commonDtos.StandardResponse{
			Status:  isSuccess,
			Message: message,
			Data:    status,
		})
		return
	}

	ctx.JSON(http.StatusNotFound, commonDtos.StandardResponse{
		Status:  false,
		Message: "Task not found in any service",
	})
}
