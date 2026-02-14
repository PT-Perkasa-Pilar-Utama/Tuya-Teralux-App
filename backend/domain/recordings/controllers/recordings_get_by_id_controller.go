package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/recordings/usecases"
)

type RecordingsGetByIDController struct {
	useCase usecases.GetRecordingByIDUseCase
}

func NewRecordingsGetByIDController(useCase usecases.GetRecordingByIDUseCase) *RecordingsGetByIDController {
	return &RecordingsGetByIDController{
		useCase: useCase,
	}
}

// GetRecordingByID handles GET /api/recordings/:id endpoint
func (c *RecordingsGetByIDController) GetRecordingByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.useCase.GetRecordingByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Recording not found"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}
