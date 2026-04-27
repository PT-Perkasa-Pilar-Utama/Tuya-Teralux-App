package controllers

import (
	"net/http"

	commonDtos "sensio/domain/common/dtos"
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/usecases"

	"github.com/gin-gonic/gin"
)

type AudioUploadStatusController struct {
	useCase usecases.UpdateAudioUploadStatusUseCase
}

func NewAudioUploadStatusController(useCase usecases.UpdateAudioUploadStatusUseCase) *AudioUploadStatusController {
	return &AudioUploadStatusController{
		useCase: useCase,
	}
}

func (c *AudioUploadStatusController) UpdateStatus(ctx *gin.Context) {
	var req recordings_dtos.UpdateAudioUploadStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	if err := c.useCase.UpdateStatus(ctx.Request.Context(), req); err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to update audio upload status",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Audio upload status updated successfully",
	})
}
