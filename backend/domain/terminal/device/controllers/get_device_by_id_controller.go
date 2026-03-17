package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/device/dtos"
	usecases "sensio/domain/terminal/device/usecases"
	"strings"

	"github.com/gin-gonic/gin"
)

// Force usage of terminal_dtos for Swagger
var _ = terminal_dtos.DeviceSingleResponseDTO{}

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
// @Summary      Get device by ID
// @Description  Retrieve device information by device ID
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Device ID"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.DeviceResponseDTO}
// @Failure      400  {object}  dtos.ValidationErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Router       /api/devices/{id} [get]
// @Security     BearerAuth
func (c *GetDeviceByIDController) GetDeviceByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if strings.TrimSpace(id) == "" || strings.Contains(id, "INVALID") {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid ID format",
			Data:    nil,
		})
		return
	}

	device, err := c.useCase.GetDeviceByID(id)
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
		Message: "Device retrieved successfully",
		Data:    device,
	})
}
