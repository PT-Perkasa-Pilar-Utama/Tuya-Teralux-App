package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceListResponseDTO{}

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
// @Summary      Get All Devices
// @Description  Retrieves all devices
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        teralux_id query string false "Filter by Teralux ID"
// @Param        page       query int    false "Page number"
// @Param        limit      query int    false "Items per page"
// @Param        per_page   query int    false "Items per page (alias for limit)"
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.DeviceListResponseDTO}
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices [get]
func (c *GetAllDevicesController) GetAllDevices(ctx *gin.Context) {
	var filter teralux_dtos.DeviceFilterDTO
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		// Ignore or handle
	}

	devices, err := c.useCase.Execute(&filter)
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
