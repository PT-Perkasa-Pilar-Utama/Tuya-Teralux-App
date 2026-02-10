package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceListResponseDTO{}

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

	pageStr := ctx.Query("page")
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limitStr = ctx.Query("per_page")
	}

	page := 0
	limit := 0
	if val, err := strconv.Atoi(pageStr); err == nil {
		page = val
	}
	if val, err := strconv.Atoi(limitStr); err == nil {
		limit = val
	}

	// Execute dedicated use case
	devices, err := c.useCase.Execute(teraluxID, page, limit)
	if err != nil {
		if strings.Contains(err.Error(), "Teralux hub not found") {
			ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
				Status:  false,
				Message: "Teralux hub not found",
				Data:    nil,
			})
			return
		}

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
