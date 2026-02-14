package controllers

import (
	"net/http"
	"strings" // Added for strings.Contains
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils" // Added for utils.ValidationError
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

// UpdateDeviceStatus handles PUT /api/devices/statuses/:deviceId/:code endpoint
func (c *UpdateDeviceStatusController) UpdateDeviceStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device ID is required",
			Data:    nil,
		})
		return
	}

	var req teralux_dtos.UpdateDeviceStatusRequestDTO
	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Data:    nil,
		})
		return
	}

	// code is now only in the body

	// Extract access token
	accessToken := ctx.MustGet("access_token").(string)

	// Execute use case
	if err := c.useCase.UpdateDeviceStatus(id, &req, accessToken); err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		// Handle Not Found
		statusCode := http.StatusInternalServerError
		errorMsg := "Internal Server Error"
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Device not found") {
			statusCode = http.StatusNotFound
			errorMsg = "Device not found"
		} else if strings.Contains(err.Error(), "Invalid status code") {
			statusCode = http.StatusNotFound
			errorMsg = "Invalid status code for this device"
		}

		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: errorMsg,
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Status updated successfully",
		Data:    nil,
	})
}
