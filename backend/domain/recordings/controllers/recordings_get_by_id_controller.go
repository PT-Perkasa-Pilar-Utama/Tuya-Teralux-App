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
// @Summary Get recording by ID
// @Description Retrieve metadata for a specific recording.
// @Tags 06. Recordings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Recording ID"
// @Success 200 {object} dtos.RecordingStandardResponse{data=dtos.RecordingResponseDto}
// @Failure 401 {object} dtos.RecordingStandardResponse
// @Failure 404 {object} dtos.RecordingStandardResponse
// @Failure 500 {object} dtos.RecordingStandardResponse
// @Router /api/recordings/{id} [get]
func (c *RecordingsGetByIDController) GetRecordingByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.useCase.GetRecordingByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Recording not found"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}
