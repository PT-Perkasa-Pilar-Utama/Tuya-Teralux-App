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
func (c *RecordingsDeleteController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.useCase.DeleteRecording(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Recording deleted successfully"})
}
