package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils" // Added for utils.ValidationError
	terminal_dtos "sensio/domain/terminal/device_status/dtos"
	usecases "sensio/domain/terminal/device_status/usecases"
	"strings" // Added for strings.Contains

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

// UpdateDeviceStatus handles PUT /api/devices/:id/status endpoint
// @Summary      Update device status
// @Description  Update a specific device status (e.g., power, temperature)
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        id       path    string                                   true  "Device ID"
// @Param        request  body    terminal_dtos.UpdateDeviceStatusRequestDTO  true  "Updated status data"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      422  {object}  dtos.StandardResponse
// @Router       /api/devices/{id}/status [put]
// @Security     BearerAuth
func (c *UpdateDeviceStatusController) UpdateDeviceStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "id", Message: "Device ID is required"},
			},
		})
		return
	}

	var req terminal_dtos.UpdateDeviceStatusRequestDTO
	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
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
