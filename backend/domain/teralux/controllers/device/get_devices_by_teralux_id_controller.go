package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// GetDevicesByTeraluxIDController handles get devices by teralux ID requests
type GetDevicesByTeraluxIDController struct {
	useCase *usecases.GetDevicesByTeraluxIDUseCase
}

// NewGetDevicesByTeraluxIDController creates a new GetDevicesByTeraluxIDController instance
func NewGetDevicesByTeraluxIDController(useCase *usecases.GetDevicesByTeraluxIDUseCase) *GetDevicesByTeraluxIDController {
	return &GetDevicesByTeraluxIDController{
		useCase: useCase,
	}
}

// GetDevicesByTeraluxID handles GET /api/devices/teralux/:teralux_id endpoint
// @Summary      Get Devices by Teralux ID
// @Description  Retrieves all devices linked to a specific teralux ID
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        teralux_id   path      string  true  "Teralux ID"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.DeviceListResponseDTO}  "Returns list of devices"
// @Failure      400  {object}  dtos.StandardResponse "Invalid Teralux ID format"
// @Failure      401  {object}  dtos.StandardResponse "Unauthorized"
// @Failure      500  {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices/teralux/{teralux_id} [get]
func (c *GetDevicesByTeraluxIDController) GetDevicesByTeraluxID(ctx *gin.Context) {
	teraluxID := ctx.Param("teralux_id")
	if strings.TrimSpace(teraluxID) == "" || strings.Contains(teraluxID, "INVALID") {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid Teralux ID format",
			Data:    nil,
		})
		return
	}

	// Execute dedicated use case
	devices, err := c.useCase.Execute(teraluxID)
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
