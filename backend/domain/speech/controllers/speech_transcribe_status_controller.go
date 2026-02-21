package controllers

import (
	"net/http"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

// SpeechTranscribeStatusController handles GET /api/speech/transcribe/:transcribe_id.
type SpeechTranscribeStatusController struct {
	statusUC usecases.GetTranscriptionStatusUseCase
}

func NewSpeechTranscribeStatusController(statusUC usecases.GetTranscriptionStatusUseCase) *SpeechTranscribeStatusController {
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
// @Success 200 {object} dtos.StandardResponse{data=dtos.AsyncTranscriptionProcessStatusResponseDTO}
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

	status, err := c.statusUC.GetTranscriptionStatus(taskID)
	if err == nil {
		ctx.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "Task status retrieved successfully",
			Data:    status,
		})
		return
	}

	ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
		Status:  false,
		Message: "Task not found in any service",
	})
}
