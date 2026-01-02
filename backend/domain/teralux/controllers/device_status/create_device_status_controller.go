package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// CreateDeviceStatusController handles create device status requests
type CreateDeviceStatusController struct {
	useCase *usecases.CreateDeviceStatusUseCase
}

// NewCreateDeviceStatusController creates a new CreateDeviceStatusController instance
func NewCreateDeviceStatusController(useCase *usecases.CreateDeviceStatusUseCase) *CreateDeviceStatusController {
	return &CreateDeviceStatusController{
		useCase: useCase,
	}
}

// CreateDeviceStatus handles POST /api/device-statuses endpoint
// @Summary      Create Device Status
// @Description  Creates a new device status
// @Tags         03. DeviceStatuses
// @Accept       json
// @Produce      json
// @Param        request  body      teralux_dtos.CreateDeviceStatusRequestDTO  true  "Create Device Status Request"
// @Success      201      {object}  dtos.StandardResponse{data=teralux_dtos.CreateDeviceStatusResponseDTO}
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses [post]
func (c *CreateDeviceStatusController) CreateDeviceStatus(ctx *gin.Context) {
	var req teralux_dtos.CreateDeviceStatusRequestDTO

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
	status, err := c.useCase.Execute(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to create device status: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Device status created successfully",
		Data:    status,
	})
}
