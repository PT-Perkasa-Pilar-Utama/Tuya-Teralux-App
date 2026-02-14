package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaSendIRCommandController handles IR device command requests
type TuyaSendIRCommandController struct {
	useCase usecases.TuyaSendIRCommandUseCase
}

// NewTuyaSendIRCommandController creates a new TuyaSendIRCommandController instance
func NewTuyaSendIRCommandController(useCase usecases.TuyaSendIRCommandUseCase) *TuyaSendIRCommandController {
	return &TuyaSendIRCommandController{
		useCase: useCase,
	}
}

// SendIRACCommand handles the request to send an IR AC command
func (ctrl *TuyaSendIRCommandController) SendIRACCommand(c *gin.Context) {
	accessToken := c.MustGet("access_token").(string)

	var req tuya_dtos.TuyaIRACCommandDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Failed to bind IR AC command: %v", err)
		c.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Data:    nil,
		})
		return
	}

	infraredID := c.Param("id")
	utils.LogDebug("TuyaSendIRCommandController.SendIRACCommand: sending to %s, remoteID: %s, code: %s", infraredID, req.RemoteID, req.Code)

	success, err := ctrl.useCase.SendIRACCommand(accessToken, infraredID, req.RemoteID, req.Code, req.Value)
	if err != nil {
		utils.LogError("SendIRACCommand failed: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	utils.LogDebug("SendIRACCommand success")
	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "IR AC Command sent successfully",
		Data:    map[string]bool{"success": success},
	})
}
