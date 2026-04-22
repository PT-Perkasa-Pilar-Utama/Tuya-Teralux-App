package controllers

import (
	"net/http"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/recordings/dtos"
	"sensio/domain/recordings/usecases"

	"github.com/gin-gonic/gin"
)

type UploadIntentController struct {
	useCase usecases.UploadIntentUseCase
}

func NewUploadIntentController(useCase usecases.UploadIntentUseCase) *UploadIntentController {
	return &UploadIntentController{
		useCase: useCase,
	}
}

// CreateUploadIntent handles POST /api/upload/intent
//
// @Summary Request a signed URL for direct S3 upload
// @Description Generates a presigned S3 URL for direct audio file upload
// @Tags upload
// @Accept json
// @Produce json
// @Param request body dtos.CreateUploadIntentRequest true "Upload intent request"
// @Success 201 {object} commonDtos.StandardResponse{data=dtos.UploadIntentResponseDTO}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 401 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse
// @Security BearerAuth
func (c *UploadIntentController) CreateUploadIntent(ctx *gin.Context) {
	var req recordings_dtos.CreateUploadIntentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	// Auth is validated via Bearer token middleware - UID available in context
	_, exists := ctx.Get("uid")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, commonDtos.StandardResponse{
			Status:  false,
			Message: "Unauthorized",
		})
		return
	}

	resp, err := c.useCase.CreateUploadIntent(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to create upload intent",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, commonDtos.StandardResponse{
		Status:  true,
		Message: "Upload intent created successfully",
		Data:    resp,
	})
}