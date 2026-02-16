package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/recordings/usecases"
)

type RecordingsCreateController struct {
	useCase usecases.SaveRecordingUseCase
}

func NewRecordingsCreateController(useCase usecases.SaveRecordingUseCase) *RecordingsCreateController {
	return &RecordingsCreateController{
		useCase: useCase,
	}
}

// CreateRecording handles POST /api/recordings endpoint
// @Summary Save a new recording
// @Description Upload an audio file and save its metadata.
// @Tags 06. Recordings
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param file formData file true "Audio file"
// @Success 201 {object} dtos.RecordingStandardResponse{data=dtos.RecordingResponseDto}
// @Failure 400 {object} dtos.RecordingStandardResponse
// @Failure 401 {object} dtos.RecordingStandardResponse
// @Failure 500 {object} dtos.RecordingStandardResponse
// @Router /api/recordings [post]
func (c *RecordingsCreateController) CreateRecording(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	result, err := c.useCase.SaveRecording(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, result)
}
