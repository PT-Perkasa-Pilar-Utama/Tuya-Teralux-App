package controllers

import (
	"net/http"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/speech/dtos"

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
// @Success 200 {object} dtos.StandardResponse{data=dtos.AsyncTranscriptionStatusDTO}
// @Failure 404 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/speech/transcribe/{transcribe_id} [get]
func (c *SpeechTranscribeStatusController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("transcribe_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
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
			message = "Task failed"
			if status.HTTPStatusCode != 0 {
				httpStatus = status.HTTPStatusCode
			}
		}

		ctx.JSON(httpStatus, dtos.StandardResponse{
			Status:  isSuccess,
			Message: message,
			Data:    status,
		})
		return
	}

	ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
		Status:  false,
		Message: "Task not found in any service",
	})
}
