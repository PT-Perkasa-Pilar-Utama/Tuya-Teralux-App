package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	common_dtos "teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	recordings_dtos "teralux_app/domain/recordings/dtos"
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
// @Success 201 {object} recordings_dtos.StandardResponse{data=recordings_dtos.RecordingResponseDto}
// @Failure 400 {object} recordings_dtos.StandardResponse
// @Failure 401 {object} recordings_dtos.StandardResponse
// @Failure 500 {object} recordings_dtos.StandardResponse "Internal Server Error"
// @Router /api/recordings [post]
func (c *RecordingsCreateController) CreateRecording(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common_dtos.StandardResponse{
			Status:  false,
			Message: "No file uploaded",
		})
		return
	}

	result, err := c.useCase.SaveRecording(file)
	if err != nil {
		utils.LogError("RecordingsCreateController.CreateRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, common_dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	// Wrap entity in DTO for response consistency if needed,
	// but here we can just use a simple map or DTO
	resp := recordings_dtos.RecordingResponseDto{
		ID:           result.ID,
		Filename:     result.Filename,
		OriginalName: result.OriginalName,
		AudioUrl:     result.AudioUrl,
		CreatedAt:    result.CreatedAt,
	}

	ctx.JSON(http.StatusCreated, common_dtos.StandardResponse{
		Status:  true,
		Message: "Recording created successfully",
		Data:    resp,
	})
}
