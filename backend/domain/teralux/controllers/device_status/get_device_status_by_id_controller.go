package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceStatusResponseDTO{}


// GetDeviceStatusByIDController handles get device status by ID requests
type GetDeviceStatusByIDController struct {
	useCase *usecases.GetDeviceStatusByIDUseCase
}

// NewGetDeviceStatusByIDController creates a new GetDeviceStatusByIDController instance
func NewGetDeviceStatusByIDController(useCase *usecases.GetDeviceStatusByIDUseCase) *GetDeviceStatusByIDController {
	return &GetDeviceStatusByIDController{
		useCase: useCase,
	}
}

// GetDeviceStatusByID handles GET /api/device-statuses/:id endpoint
// @Summary      Get Device Status by ID
// @Description  Retrieves a single device status by ID
// @Tags         03. DeviceStatuses
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device Status ID"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.DeviceStatusResponseDTO}
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/device-statuses/{id} [get]
func (c *GetDeviceStatusByIDController) GetDeviceStatusByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Device Status ID is required",
			Data:    nil,
		})
		return
	}

	status, err := c.useCase.Execute(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to retrieve device status: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device status retrieved successfully",
		Data:    status,
	})
}
