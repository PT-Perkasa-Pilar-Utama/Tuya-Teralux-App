package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	tuya_dtos "sensio/domain/tuya/dtos"
	"sensio/domain/tuya/usecases"

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
// @Summary      Send Switch Command
// @Description  Sends a standard switch command (e.g., toggle power) to a specific Tuya device.
// @Tags 01. Tuya
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Device ID"
// @Param        body  body      tuya_dtos.TuyaCommandDTO  true  "Command Payload"
// @Success      200   {object}  dtos.StandardResponse
// @Failure      400   {object}  dtos.StandardResponse
// @Failure      500   {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices/{id}/commands/switch [post]
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
