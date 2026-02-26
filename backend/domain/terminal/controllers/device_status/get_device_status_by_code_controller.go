package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/dtos"
	usecases "sensio/domain/terminal/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// Force usage of terminal_dtos for Swagger
var _ = terminal_dtos.DeviceStatusSingleResponseDTO{}

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
func (c *GetDeviceStatusByCodeController) GetDeviceStatusByCode(ctx *gin.Context) {
	deviceID := ctx.Param("id")
	code := ctx.Param("code")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "id", Message: "Device ID is required"},
			},
		})
		return
	}
	if code == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "code", Message: "Status code is required"},
			},
		})
		return
	}

	status, err := c.useCase.GetDeviceStatusByCode(deviceID, code)
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
