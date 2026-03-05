package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/tasks"
	"sensio/domain/speech/dtos"

	"github.com/gin-gonic/gin"
)

// SpeechTranscribeStatusController handles GET /api/speech/transcribe/:transcribe_id.
type SpeechTranscribeStatusController struct {
	statusUC tasks.GenericStatusUseCase[dtos.AsyncTranscriptionStatusDTO]
}

func NewSpeechTranscribeStatusController(statusUC tasks.GenericStatusUseCase[dtos.AsyncTranscriptionStatusDTO]) *SpeechTranscribeStatusController {
	return &SpeechTranscribeStatusController{
		statusUC: statusUC,
	}
}

// GetStatus handles GET /api/speech/transcribe/:transcribe_id
// @Summary Get transcription status (Consolidated)
// @Description Get the status and result of any transcription task (Short, Long, or Orion).
// @Tags 04. Speech
// @Security BearerAuth
// @Produce json
// @Param transcribe_id path string true "Task ID"
// @Success 200 {object} commonDtos.StandardResponse{data=dtos.AsyncTranscriptionStatusDTO}
// @Failure 404 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/speech/transcribe/{transcribe_id} [get]
func (c *SpeechTranscribeStatusController) GetStatus(ctx *gin.Context) {
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
