package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceStatusSingleResponseDTO{}

// GetDeviceStatusByCodeController handles get device status by code requests
type GetDeviceStatusByCodeController struct {
	useCase *usecases.GetDeviceStatusByCodeUseCase
}

// NewGetDeviceStatusByCodeController creates a new GetDeviceStatusByCodeController instance
func NewGetDeviceStatusByCodeController(useCase *usecases.GetDeviceStatusByCodeUseCase) *GetDeviceStatusByCodeController {
	return &GetDeviceStatusByCodeController{
		useCase: useCase,
	}
}

// GetDeviceStatusByCode handles GET /api/devices/:id/statuses/:code endpoint
// @Summary      Get Device Status by Code
// @Tags         05. Device Statuses
// @Produce      json
// @Param        id        path      string  true  "Device ID"
// @Param        code      path      string  true  "Status Code"
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.DeviceStatusSingleResponseDTO}
// @Failure      400      {object}  dtos.StandardResponse "Validation Error: id is required"
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      404      {object}  dtos.StandardResponse "Status code not found for this device"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices/{id}/statuses/{code} [get]
func (c *GetDeviceStatusByCodeController) GetDeviceStatusByCode(ctx *gin.Context) {
	deviceID := ctx.Param("id")
	code := ctx.Param("code")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error: id is required",
			Data:    nil,
		})
		return
	}
	if code == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error: code is required",
			Data:    nil,
		})
		return
	}

	status, err := c.useCase.Execute(deviceID, code)
	if err != nil {
		errorMsg := "Status code not found for this device"
		if err.Error() == "Device not found" {
			errorMsg = "Device not found"
		}
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: errorMsg,
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Status retrieved successfully",
		Data:    status,
	})
}
