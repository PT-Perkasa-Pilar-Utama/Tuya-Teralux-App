package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// DeleteDeviceStatusController handles delete device status requests
type DeleteDeviceStatusController struct {
	useCase *usecases.DeleteDeviceStatusUseCase
}

// NewDeleteDeviceStatusController creates a new DeleteDeviceStatusController instance
func NewDeleteDeviceStatusController(useCase *usecases.DeleteDeviceStatusUseCase) *DeleteDeviceStatusController {
	return &DeleteDeviceStatusController{
		useCase: useCase,
	}
}

// DeleteDeviceStatus handles DELETE /api/device-statuses/:id endpoint
// @Summary      Delete Device Status
// @Description  Deletes a device status
// @Tags         05. Device Statuses
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device Status ID"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses/{id} [delete]
func (c *DeleteDeviceStatusController) DeleteDeviceStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device Status ID is required",
			Data:    nil,
		})
		return
	}

	if err := c.useCase.Execute(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to delete device status: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device status deleted successfully",
		Data:    nil,
	})
}
