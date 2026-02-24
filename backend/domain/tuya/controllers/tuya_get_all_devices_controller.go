package controllers

import (
	"net/http"
	"strconv"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = tuya_dtos.TuyaDevicesResponseDTO{}

type TuyaGetAllDevicesController struct {
	useCase usecases.TuyaGetAllDevicesUseCase
}

// NewTuyaGetAllDevicesController creates a new TuyaGetAllDevicesController instance
func NewTuyaGetAllDevicesController(useCase usecases.TuyaGetAllDevicesUseCase) *TuyaGetAllDevicesController {
	return &TuyaGetAllDevicesController{
		useCase: useCase,
	}
}

// GetAllDevices handles GET /api/tuya/devices endpoint
// @Summary      Get All Devices
// @Description  Retrieves a list of all devices in a Merged View (Smart IR remotes merged with Hubs). Sorted alphabetically by Name. For infrared_ac devices, the status array is populated with saved device state (power, temp, mode, wind) or default values if no state exists.
// @Tags         02. Tuya
// @Accept       json
// @Produce      json
// @Param        page      query  int     false  "Page number"
// @Param        limit     query  int     false  "Items per page"
// @Param        per_page  query  int     false  "Items per page (alias for limit)"
// @Param        category  query  string  false  "Filter by category"
// @Success      200  {object}  dtos.StandardResponse{data=tuya_dtos.TuyaDevicesResponseDTO}
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices [get]
func (c *TuyaGetAllDevicesController) GetAllDevices(ctx *gin.Context) {
	accessToken := ctx.MustGet("access_token").(string)

	uid := utils.AppConfig.TuyaUserID
	if uid == "" {
		utils.LogError("TUYA_USER_ID is not set in environment")
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Server configuration error: TUYA_USER_ID missing",
			Data:    nil,
		})
		return
	}
	utils.LogDebug("Using TUYA_USER_ID from env: '%s'", uid)

	// Parse optional query parameters
	pageStr := ctx.Query("page")
	limitStr := ctx.Query("limit")
	// Support per_page alias
	if limitStr == "" {
		limitStr = ctx.Query("per_page")
	}
	category := ctx.Query("category")

	page := 0
	limit := 0
	var errConv error

	if pageStr != "" {
		page, errConv = strconv.Atoi(pageStr)
		if errConv != nil {
			utils.LogWarn("Invalid page parameter: %v", errConv)
			page = 0 // Default to 0 (ignored)
		}
	}

	if limitStr != "" {
		limit, errConv = strconv.Atoi(limitStr)
		if errConv != nil {
			utils.LogWarn("Invalid limit parameter: %v", errConv)
			limit = 0 // Default to 0 (ignored)
		}
	}

	devices, err := c.useCase.GetAllDevices(accessToken, uid, page, limit, category)
	if err != nil {
		utils.LogError("TuyaGetAllDevicesController.GetAllDevices: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Devices fetched successfully",
		Data:    devices,
	})
}
