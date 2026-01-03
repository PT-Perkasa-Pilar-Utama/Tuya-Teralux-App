package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// UpdateDeviceStatusController handles update device status requests
type UpdateDeviceStatusController struct {
	useCase *usecases.UpdateDeviceStatusUseCase
}

// NewUpdateDeviceStatusController creates a new UpdateDeviceStatusController instance
func NewUpdateDeviceStatusController(useCase *usecases.UpdateDeviceStatusUseCase) *UpdateDeviceStatusController {
	return &UpdateDeviceStatusController{
		useCase: useCase,
	}
}

// UpdateDeviceStatus handles PUT /api/device-statuses/:deviceId/:code endpoint
// @Summary      Update Device Status
// @Description  Updates an existing device status
// @Tags         05. Device Statuses
// @Accept       json
// @Produce      json
// @Param        deviceId path      string                              true  "Device ID"
// @Param        code     path      string                              true  "Status Code"
// @Param        request  body      teralux_dtos.UpdateDeviceStatusRequestDTO  true  "Update Device Status Request"
// @Success      200      {object}  dtos.StandardResponse
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses/{deviceId}/{code} [put]
func (c *UpdateDeviceStatusController) UpdateDeviceStatus(ctx *gin.Context) {
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

	var req teralux_dtos.UpdateDeviceStatusRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body: " + err.Error(),
			Data:    nil,
		})
		return
	}

	if err := c.useCase.Execute(deviceID, code, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to update device status: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device status updated successfully",
		Data:    nil,
	})
}
