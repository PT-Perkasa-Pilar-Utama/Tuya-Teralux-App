package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/recordings/usecases"
)

// Force usage of commonDtos for Swagger
var _ = commonDtos.StandardResponse{}

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
// @Success 200 {object} commonDtos.StandardResponse
// @Failure      401  {object}  commonDtos.ErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router /api/recordings/{id} [delete]
func (c *RecordingsDeleteController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.useCase.DeleteRecording(id)
	if err != nil {
		utils.LogError("RecordingsDeleteController.DeleteRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}
	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Recording deleted successfully",
	})
}
