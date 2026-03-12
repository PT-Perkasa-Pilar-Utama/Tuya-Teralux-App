package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/usecases"
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
// @Param mac_address formData string false "Device MAC Address"
// @Success 201 {object} commonDtos.StandardResponse{data=recordings_dtos.RecordingResponseDto}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 401 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/recordings [post]
func (c *RecordingsCreateController) CreateRecording(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "No file uploaded",
		})
		return
	}

	macAddress := ctx.PostForm("mac_address")

	baseURL := utils.GetBaseURL(ctx)
	result, err := c.useCase.SaveRecording(file, macAddress, baseURL)
	if err != nil {
		utils.LogError("RecordingsCreateController.CreateRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
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

	ctx.JSON(http.StatusCreated, commonDtos.StandardResponse{
		Status:  true,
		Message: "Recording created successfully",
		Data:    resp,
	})
}
