package controllers

import (
	"net/http"

	common_dtos "teralux_app/domain/common/dtos"
	recordings_dtos "teralux_app/domain/recordings/dtos"
	"teralux_app/domain/recordings/usecases"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = recordings_dtos.RecordingResponseDto{}

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
// @Tags 07. Recordings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Recording ID"
// @Success 200 {object} recordings_dtos.StandardResponse{data=recordings_dtos.RecordingResponseDto}
// @Failure 401 {object} recordings_dtos.StandardResponse
// @Failure 404 {object} recordings_dtos.StandardResponse
// @Failure 500 {object} recordings_dtos.StandardResponse "Internal Server Error"
// @Router /api/recordings/{id} [get]
func (c *RecordingsGetByIDController) GetRecordingByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.useCase.GetRecordingByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, common_dtos.StandardResponse{
			Status:  false,
			Message: "Recording not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, common_dtos.StandardResponse{
		Status:  true,
		Message: "Recording retrieved successfully",
		Data:    result,
	})
}
