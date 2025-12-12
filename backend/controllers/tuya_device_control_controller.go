package controllers

import (
	"log"
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaDeviceControlController handles device control requests
type TuyaDeviceControlController struct {
	useCase *usecases.TuyaDeviceControlUseCase
}

// NewTuyaDeviceControlController creates a new TuyaDeviceControlController instance
func NewTuyaDeviceControlController(useCase *usecases.TuyaDeviceControlUseCase) *TuyaDeviceControlController {
	return &TuyaDeviceControlController{
		useCase: useCase,
	}
}

// SendCommand handles the request to send commands to a device
func (ctrl *TuyaDeviceControlController) SendCommand(c *gin.Context) {
	deviceID := c.Param("id")
	accessToken := c.GetHeader("access_token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing access_token header"})
		return
	}

	var req dtos.TuyaCommandsRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, err := ctrl.useCase.SendCommand(accessToken, deviceID, req.Commands)
	if err != nil {
		log.Printf("ERROR: SendCommand failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": success})
}

// SendIRACCommand handles the request to send a command to an IR air conditioner
func (ctrl *TuyaDeviceControlController) SendIRACCommand(c *gin.Context) {
	accessToken := c.GetHeader("access_token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing access_token header"})
		return
	}

	var req dtos.TuyaIRACCommandDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: Failed to bind IR AC command: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, err := ctrl.useCase.SendIRACCommand(accessToken, req.InfraredID, req.RemoteID, req.Code, req.Value)
	if err != nil {
		log.Printf("ERROR: SendIRACCommand failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": success})
}
