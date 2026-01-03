package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceStatusResponseDTO{}

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

// GetDeviceStatusByCode handles GET /api/device-statuses/:deviceId/:code endpoint
// @Summary      Get Device Status by Code
// @Description  Retrieves a single device status by device ID and code
// @Tags         05. Device Statuses
// @Produce      json
// @Param        deviceId path      string  true  "Device ID"
// @Param        code     path      string  true  "Status Code"
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.DeviceStatusResponseDTO}
// @Failure      404      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses/{deviceId}/{code} [get]
func (c *GetDeviceStatusByCodeController) GetDeviceStatusByCode(ctx *gin.Context) {
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

	status, err := c.useCase.Execute(deviceID, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to get device status: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device status retrieved successfully",
		Data:    status,
	})
}
