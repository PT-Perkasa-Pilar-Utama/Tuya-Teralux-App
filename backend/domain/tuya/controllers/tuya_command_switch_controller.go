package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaCommandSwitchController handles switch command requests
type TuyaCommandSwitchController struct {
	useCase usecases.TuyaCommandSwitchUseCase
}

// NewTuyaCommandSwitchController creates a new TuyaCommandSwitchController instance
func NewTuyaCommandSwitchController(useCase usecases.TuyaCommandSwitchUseCase) *TuyaCommandSwitchController {
	return &TuyaCommandSwitchController{
		useCase: useCase,
	}
}

// SendSwitchCommand handles the request to send switch commands to a device
func (ctrl *TuyaCommandSwitchController) SendSwitchCommand(c *gin.Context) {
	deviceID := c.Param("id")
	accessToken := c.MustGet("access_token").(string)
	utils.LogDebug("TuyaCommandSwitchController.SendSwitchCommand: received request for device %s", deviceID)

	var req tuya_dtos.TuyaCommandDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Failed to bind command: %v", err)
		c.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	commands := []tuya_dtos.TuyaCommandDTO{req}
	success, err := ctrl.useCase.SendSwitchCommand(accessToken, deviceID, commands)
	if err != nil {
		utils.LogError("TuyaCommandSwitchController.SendSwitchCommand: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	utils.LogDebug("SendSwitchCommand success")
	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Command sent successfully",
		Data:    map[string]bool{"success": success},
	})
}
