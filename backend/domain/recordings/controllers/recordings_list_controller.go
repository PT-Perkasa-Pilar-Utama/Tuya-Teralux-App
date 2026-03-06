package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/usecases"
)

// Force import for Swagger
var _ = recordings_dtos.RecordingResponseDto{}

type RecordingsListController struct {
	useCase usecases.GetAllRecordingsUseCase
}

func NewRecordingsListController(useCase usecases.GetAllRecordingsUseCase) *RecordingsListController {
	return &RecordingsListController{
		useCase: useCase,
	}
}

// ListRecordings handles GET /api/recordings endpoint
// @Summary List all recordings
// @Description Get a paginated list of all recordings.
// @Tags 07. Recordings
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} commonDtos.StandardResponse{data=recordings_dtos.GetAllRecordingsResponseDto}
// @Failure 401 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/recordings [get]
func (c *RecordingsListController) ListRecordings(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	result, err := c.useCase.ListRecordings(page, limit)
	if err != nil {
		utils.LogError("RecordingsListController.ListRecordings: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Recordings retrieved successfully",
		Data:    result,
	})
}
