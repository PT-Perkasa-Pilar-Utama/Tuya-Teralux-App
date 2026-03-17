package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/device/dtos"
	usecases "sensio/domain/terminal/device/usecases"

	"github.com/gin-gonic/gin"
)

// Force usage of terminal_dtos for Swagger
var _ = terminal_dtos.DeviceListResponseDTO{}

// GetAllDevicesController handles get all devices requests
type GetAllDevicesController struct {
	useCase *usecases.GetAllDevicesUseCase
}

// NewGetAllDevicesController creates a new GetAllDevicesController instance
func NewGetAllDevicesController(useCase *usecases.GetAllDevicesUseCase) *GetAllDevicesController {
	return &GetAllDevicesController{
		useCase: useCase,
	}
}
// GetAllDevices handles GET /api/devices endpoint
// @Summary      Get all devices
// @Description  Retrieve a list of all registered devices with optional filtering
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        terminal_id  query    string  false  "Filter by terminal ID"
// @Param        name         query    string  false  "Filter by device name"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.DeviceListResponseDTO}
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /api/devices [get]
// @Security     BearerAuth

func (c *GetAllDevicesController) GetAllDevices(ctx *gin.Context) {
	var filter terminal_dtos.DeviceFilterDTO
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		utils.LogWarn("GetAllDevices: Failed to bind query filter: %v", err)
	}

	devices, err := c.useCase.ListDevices(&filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Devices retrieved successfully",
		Data:    devices,
	})
}
