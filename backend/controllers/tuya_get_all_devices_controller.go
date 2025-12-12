package controllers

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"

	"log"

	"github.com/gin-gonic/gin"
)

// TuyaGetAllDevicesController handles get all devices requests for Tuya
type TuyaGetAllDevicesController struct {
	useCase *usecases.TuyaGetAllDevicesUseCase
}

// NewTuyaGetAllDevicesController creates a new TuyaGetAllDevicesController instance
func NewTuyaGetAllDevicesController(useCase *usecases.TuyaGetAllDevicesUseCase) *TuyaGetAllDevicesController {
	return &TuyaGetAllDevicesController{
		useCase: useCase,
	}
}

// GetAllDevices handles GET /api/tuya/devices endpoint
func (c *TuyaGetAllDevicesController) GetAllDevices(ctx *gin.Context) {
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
	uid := ctx.GetHeader("tuya_uid")
	log.Printf("DEBUG: Received tuya_uid header: '%s'", uid)
	if uid == "" {
		errorResponse := dtos.ErrorResponseDTO{
			Error:   "Missing user ID",
			Message: "tuya_uid header is required",
		}
		ctx.JSON(http.StatusUnauthorized, errorResponse)
		return
	}

	devices, err := c.useCase.GetAllDevices(accessToken, uid)
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		errorResponse := dtos.ErrorResponseDTO{
			Error:   "Failed to fetch devices",
			Message: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Return success response
	ctx.JSON(http.StatusOK, devices)
}
