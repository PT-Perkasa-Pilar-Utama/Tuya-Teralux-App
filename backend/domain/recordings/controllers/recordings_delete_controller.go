package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/recordings/usecases"
)

type RecordingsDeleteController struct {
	useCase usecases.DeleteRecordingUseCase
}

func NewRecordingsDeleteController(useCase usecases.DeleteRecordingUseCase) *RecordingsDeleteController {
	return &RecordingsDeleteController{
		useCase: useCase,
	}
}

// DeleteRecording handles DELETE /api/recordings/:id endpoint
// @Summary Delete a recording
// @Description Remove a recording and its associated file from the system.
// @Tags 06. Recordings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Recording ID"
// @Success 200 {object} dtos.RecordingStandardResponse
// @Failure 401 {object} dtos.RecordingStandardResponse
// @Failure 500 {object} dtos.RecordingStandardResponse
// @Router /api/recordings/{id} [delete]
func (c *RecordingsDeleteController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.useCase.DeleteRecording(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Recording deleted successfully"})
}
