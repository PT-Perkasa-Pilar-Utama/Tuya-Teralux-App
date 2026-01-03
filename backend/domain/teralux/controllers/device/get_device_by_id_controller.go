package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceResponseDTO{}

// GetDeviceByIDController handles get device by ID requests
type GetDeviceByIDController struct {
	useCase *usecases.GetDeviceByIDUseCase
}

// NewGetDeviceByIDController creates a new GetDeviceByIDController instance
func NewGetDeviceByIDController(useCase *usecases.GetDeviceByIDUseCase) *GetDeviceByIDController {
	return &GetDeviceByIDController{
		useCase: useCase,
	}
}

// GetDeviceByID handles GET /api/devices/:id endpoint
// @Summary      Get Device by ID
// @Description  Retrieves a single device by ID
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device ID"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.DeviceResponseDTO}
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/devices/{id} [get]
func (c *GetDeviceByIDController) GetDeviceByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device ID is required",
			Data:    nil,
		})
		return
	}

	device, err := c.useCase.Execute(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to retrieve device: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device retrieved successfully",
		Data:    device,
	})
}
