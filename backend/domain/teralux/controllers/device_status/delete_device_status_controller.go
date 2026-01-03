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

// DeleteDeviceStatus handles DELETE /api/device-statuses/:deviceId/:code endpoint
// @Summary      Delete Device Status
// @Description  Deletes an existing device status
// @Tags         05. Device Statuses
// @Param        deviceId path      string  true  "Device ID"
// @Param        code     path      string  true  "Status Code"
// @Success      200      {object}  dtos.StandardResponse
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses/{deviceId}/{code} [delete]
func (c *DeleteDeviceStatusController) DeleteDeviceStatus(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	code := ctx.Param("code")
	if deviceID == "" || code == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device ID and Status Code are required",
			Data:    nil,
		})
		return
	}

	if err := c.useCase.Execute(deviceID, code); err != nil {
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
