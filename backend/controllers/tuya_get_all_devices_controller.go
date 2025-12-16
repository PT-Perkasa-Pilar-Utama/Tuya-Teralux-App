package controllers

import (
	"log"
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"
	"teralux_app/utils"

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
	// Get access token from context (set by middleware)
	accessToken := ctx.MustGet("access_token").(string)

	// Use Tuya User ID from environment/config
	uid := utils.AppConfig.TuyaUserID
	if uid == "" {
		log.Println("ERROR: TUYA_USER_ID is not set in environment")
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Server configuration error: TUYA_USER_ID missing",
			Data:    nil,
		})
		return
	}
	log.Printf("DEBUG: Using TUYA_USER_ID from env: '%s'", uid)

	devices, err := c.useCase.GetAllDevices(accessToken, uid)
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Return success response
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Devices fetched successfully",
		Data:    devices,
	})
}
