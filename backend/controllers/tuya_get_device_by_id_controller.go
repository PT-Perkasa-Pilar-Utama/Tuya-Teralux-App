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
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "device ID is required",
			Data:    nil,
		})
		return
	}

	// Get access token from context (set by middleware)
	accessToken := ctx.MustGet("access_token").(string)

	// Call use case
	device, err := c.useCase.GetDeviceByID(accessToken, deviceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Return success response
	// Return success response
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device fetched successfully",
		Data:    dtos.TuyaDeviceResponseDTO{Device: *device},
	})
}
