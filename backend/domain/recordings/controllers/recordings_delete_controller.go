package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	common_dtos "teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	recordings_dtos "teralux_app/domain/recordings/dtos"
	"teralux_app/domain/recordings/usecases"
)

// Force import for Swagger
var _ = recordings_dtos.StandardResponse{}

type RecordingsDeleteController struct {
	useCase usecases.DeleteRecordingUseCase
}

func NewRecordingsDeleteController(useCase usecases.DeleteRecordingUseCase) *RecordingsDeleteController {
	return &RecordingsDeleteController{
		useCase: useCase,
	}
}

// DeleteRecording handles DELETE /api/recordings/:id endpoint
// @Summary Delete a recording
// @Description Remove a recording and its associated file from the system.
// @Tags 06. Recordings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Recording ID"
// @Success 200 {object} recordings_dtos.StandardResponse
// @Failure 401 {object} recordings_dtos.StandardResponse
// @Failure 500 {object} recordings_dtos.StandardResponse "Internal Server Error"
// @Router /api/recordings/{id} [delete]
func (c *RecordingsDeleteController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.useCase.DeleteRecording(id)
	if err != nil {
		utils.LogError("RecordingsDeleteController.DeleteRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, common_dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}
	ctx.JSON(http.StatusOK, common_dtos.StandardResponse{
		Status:  true,
		Message: "Recording deleted successfully",
	})
}
