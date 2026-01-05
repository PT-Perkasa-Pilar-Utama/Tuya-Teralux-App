package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// DeleteDeviceController handles delete device requests
type DeleteDeviceController struct {
	useCase *usecases.DeleteDeviceUseCase
}

// NewDeleteDeviceController creates a new DeleteDeviceController instance
func NewDeleteDeviceController(useCase *usecases.DeleteDeviceUseCase) *DeleteDeviceController {
	return &DeleteDeviceController{
		useCase: useCase,
	}
}

// DeleteDevice handles DELETE /api/devices/:id endpoint
// @Summary      Delete Device
// @Description  Deletes a device
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device ID"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      404      {object}  dtos.StandardResponse "Device not found"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices/{id} [delete]
func (c *DeleteDeviceController) DeleteDevice(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Data:    nil,
		})
		return
	}

	if err := c.useCase.Execute(id); err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Device not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device deleted successfully",
		Data:    nil,
	})
}
