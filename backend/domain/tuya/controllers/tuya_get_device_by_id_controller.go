package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	tuya_dtos "sensio/domain/tuya/dtos"
	"sensio/domain/tuya/usecases"

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
// @Summary      Get Device by ID
// @Description  Retrieves details of a specific device by its ID
// @Tags 01. Tuya
// @Accept       json
// @Produce      json
// @Param        id         path      string  true   "Device ID"
// @Param        remote_id  query     string  false  "Optional Remote sub-device ID"
// @Success      200  {object}  dtos.StandardResponse{data=tuya_dtos.TuyaDeviceResponseDTO}
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices/{id} [get]
func (c *TuyaGetDeviceByIDController) GetDeviceByID(ctx *gin.Context) {
	deviceID := ctx.Param("id")
	remoteID := ctx.Query("remote_id")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "id", Message: "Device ID is required"},
			},
		})
		return
	}

	accessToken := ctx.MustGet("access_token").(string)
	utils.LogDebug("GetDeviceByID: requesting device %s (remote=%s)", deviceID, remoteID)
	device, err := c.useCase.GetDeviceByID(accessToken, deviceID, remoteID)
	if err != nil {
		utils.LogError("TuyaGetDeviceByIDController.GetDeviceByID: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	utils.LogDebug("GetDeviceByID success")
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device fetched successfully",
		Data:    tuya_dtos.TuyaDeviceResponseDTO{Device: *device},
	})
}
