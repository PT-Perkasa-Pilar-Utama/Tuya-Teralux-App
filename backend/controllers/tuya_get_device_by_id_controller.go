package controllers

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaGetDeviceByIDController handles get device by ID requests for Tuya
type TuyaGetDeviceByIDController struct {
	useCase *usecases.TuyaGetDeviceByIDUseCase
}

// NewTuyaGetDeviceByIDController creates a new TuyaGetDeviceByIDController instance
func NewTuyaGetDeviceByIDController(useCase *usecases.TuyaGetDeviceByIDUseCase) *TuyaGetDeviceByIDController {
	return &TuyaGetDeviceByIDController{
		useCase: useCase,
	}
}

// GetDeviceByID handles GET /api/tuya/devices/:id endpoint
func (c *TuyaGetDeviceByIDController) GetDeviceByID(ctx *gin.Context) {
	// Get device ID from URL parameter
	deviceID := ctx.Param("id")
	if deviceID == "" {
		errorResponse := dtos.ErrorResponseDTO{
			Error:   "Missing device ID",
			Message: "device ID is required",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Get access token from header
	accessToken := ctx.GetHeader("access_token")
	if accessToken == "" {
		errorResponse := dtos.ErrorResponseDTO{
			Error:   "Missing access token",
			Message: "access_token header is required",
		}
		ctx.JSON(http.StatusUnauthorized, errorResponse)
		return
	}

	// Call use case
	device, err := c.useCase.GetDeviceByID(accessToken, deviceID)
	if err != nil {
		errorResponse := dtos.ErrorResponseDTO{
			Error:   "Failed to fetch device",
			Message: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Return success response
	response := dtos.TuyaDeviceResponseDTO{
		Device: *device,
	}
	ctx.JSON(http.StatusOK, response)
}
