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

// UpdateDeviceStatus handles PUT /api/device-statuses/:id endpoint
// @Summary      Update Device Status
// @Description  Updates an existing device status
// @Tags         05. Device Statuses
// @Accept       json
// @Produce      json
// @Param        id       path      string                              true  "Device Status ID"
// @Param        request  body      teralux_dtos.UpdateDeviceStatusRequestDTO  true  "Update Device Status Request"
// @Success      200      {object}  dtos.StandardResponse
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses/{id} [put]
func (c *UpdateDeviceStatusController) UpdateDeviceStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device Status ID is required",
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

	if err := c.useCase.Execute(id, &req); err != nil {
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
