package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceStatusListResponseDTO{}


// GetAllDeviceStatusesController handles get all device statuses requests
type GetAllDeviceStatusesController struct {
	useCase *usecases.GetAllDeviceStatusesUseCase
}

// NewGetAllDeviceStatusesController creates a new GetAllDeviceStatusesController instance
func NewGetAllDeviceStatusesController(useCase *usecases.GetAllDeviceStatusesUseCase) *GetAllDeviceStatusesController {
	return &GetAllDeviceStatusesController{
		useCase: useCase,
	}
}

// GetAllDeviceStatuses handles GET /api/device-statuses endpoint
// @Summary      Get All Device Statuses
// @Description  Retrieves all device statuses
// @Tags         03. DeviceStatuses
// @Accept       json
// @Produce      json
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.DeviceStatusListResponseDTO}
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses [get]
func (c *GetAllDeviceStatusesController) GetAllDeviceStatuses(ctx *gin.Context) {
	statuses, err := c.useCase.Execute()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to retrieve device statuses: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device statuses retrieved successfully",
		Data:    statuses,
	})
}
