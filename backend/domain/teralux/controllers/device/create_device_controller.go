package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// CreateDeviceController handles create device requests
type CreateDeviceController struct {
	useCase *usecases.CreateDeviceUseCase
}

// NewCreateDeviceController creates a new CreateDeviceController instance
func NewCreateDeviceController(useCase *usecases.CreateDeviceUseCase) *CreateDeviceController {
	return &CreateDeviceController{
		useCase: useCase,
	}
}

// CreateDevice handles POST /api/devices endpoint
// @Summary      Create Device
// @Description  Creates a new device under a teralux unit
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        request  body      teralux_dtos.CreateDeviceRequestDTO  true  "Create Device Request"
// @Success      201      {object}  dtos.StandardResponse{data=teralux_dtos.CreateDeviceResponseDTO}
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/devices [post]
func (c *CreateDeviceController) CreateDevice(ctx *gin.Context) {
	var req teralux_dtos.CreateDeviceRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Execute use case
	device, err := c.useCase.Execute(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to create device: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Device created successfully",
		Data:    device,
	})
}
