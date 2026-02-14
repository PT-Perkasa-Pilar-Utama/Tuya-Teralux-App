package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// SyncDeviceStatusController handles device status synchronization requests
type SyncDeviceStatusController struct {
	useCase *usecases.SyncDeviceStatusUseCase
}

// NewSyncDeviceStatusController creates a new instance of SyncDeviceStatusController
func NewSyncDeviceStatusController(useCase *usecases.SyncDeviceStatusUseCase) *SyncDeviceStatusController {
	return &SyncDeviceStatusController{
		useCase: useCase,
	}
}

// SyncStatus handles the request to synchronize Teralux device status with Tuya
// @Summary Sync Teralux Device Status
// @Description Fetches real-time status from Tuya. Does NOT update local DB. Returns fresh device list.
// @Tags 02. Tuya
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Success 200 {object} dtos.StandardResponse{data=[]dtos.TuyaSyncDeviceDTO}
// @Failure 401 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/tuya/devices/sync [get]
func (ctrl *SyncDeviceStatusController) SyncStatus(c *gin.Context) {
	// 1. Extract Access Token from Context (set by AuthMiddleware/TuyaMiddleware)
	val, exists := c.Get("access_token")
	if !exists {
		utils.LogError("access_token missing in context")
		c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
			Status:  false,
			Message: "Unauthorized: Access token missing",
			Data:    nil,
		})
		return
	}
	accessToken := val.(string)

	// 2. Get Tuya User ID from Config
	uid := utils.AppConfig.TuyaUserID
	if uid == "" {
		utils.LogError("TUYA_USER_ID is not set in environment")
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Server configuration error: TUYA_USER_ID missing",
			Data:    nil,
		})
		return
	}

	// 3. Execute Sync
	resp, err := ctrl.useCase.SyncDeviceStatuses(accessToken, uid)
	if err != nil {
		utils.LogError("Failed to sync device status: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to sync device status: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device status synced successfully",
		Data:    resp,
	})
}
