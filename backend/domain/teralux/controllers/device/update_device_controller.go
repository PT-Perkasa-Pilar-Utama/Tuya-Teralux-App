package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// UpdateDeviceController handles update device requests
type UpdateDeviceController struct {
	useCase *usecases.UpdateDeviceUseCase
}

// NewUpdateDeviceController creates a new UpdateDeviceController instance
func NewUpdateDeviceController(useCase *usecases.UpdateDeviceUseCase) *UpdateDeviceController {
	return &UpdateDeviceController{
		useCase: useCase,
	}
}

// UpdateDevice handles PUT /api/devices/:id endpoint
// @Summary      Update Device
// @Description  Updates an existing device
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Device ID"
// @Param        request  body      teralux_dtos.UpdateDeviceRequestDTO  true  "Update Device Request"
// @Success      200      {object}  dtos.StandardResponse
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/devices/{id} [put]
func (c *UpdateDeviceController) UpdateDevice(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device ID is required",
			Data:    nil,
		})
		return
	}

	var req teralux_dtos.UpdateDeviceRequestDTO
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
			Message: "Failed to update device: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device updated successfully",
		Data:    nil,
	})
}
