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

// GetDeviceStatusesByDeviceIDController handles get device statuses by device ID requests
type GetDeviceStatusesByDeviceIDController struct {
	useCase *usecases.GetDeviceStatusesByDeviceIDUseCase
}

// NewGetDeviceStatusesByDeviceIDController creates a new GetDeviceStatusesByDeviceIDController instance
func NewGetDeviceStatusesByDeviceIDController(useCase *usecases.GetDeviceStatusesByDeviceIDUseCase) *GetDeviceStatusesByDeviceIDController {
	return &GetDeviceStatusesByDeviceIDController{
		useCase: useCase,
	}
}

// GetDeviceStatusesByDeviceID handles GET /api/device/statuses/:deviceId endpoint
// @Summary      Get Device Statuses by Device ID
// @Description  Retrieves all statuses for a specific device
// @Tags         05. Device Statuses
// @Produce      json
// @Param        id   path      string  true  "Device ID"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.DeviceStatusListResponseDTO}
// @Failure      401  {object}  dtos.StandardResponse "Unauthorized"
// @Failure      404  {object}  dtos.StandardResponse "Device not found"
// @Failure      500  {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices/{id}/statuses [get]
func (c *GetDeviceStatusesByDeviceIDController) GetDeviceStatusesByDeviceID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device ID is required",
			Data:    nil,
		})
		return
	}

	statuses, err := c.useCase.Execute(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Device not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Statuses retrieved successfully",
		Data:    statuses,
	})
}
