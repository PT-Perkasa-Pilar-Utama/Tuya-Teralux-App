package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/dtos"
	usecases "sensio/domain/terminal/usecases/device"

	"github.com/gin-gonic/gin"
)

// Force usage of terminal_dtos for Swagger
var _ = terminal_dtos.DeviceListResponseDTO{}

// GetDevicesByTerminalIDController handles get devices by terminal ID requests
type GetDevicesByTerminalIDController struct {
	useCase *usecases.GetDevicesByTerminalIDUseCase
}

// NewGetDevicesByTerminalIDController creates a new GetDevicesByTerminalIDController instance
func NewGetDevicesByTerminalIDController(useCase *usecases.GetDevicesByTerminalIDUseCase) *GetDevicesByTerminalIDController {
	return &GetDevicesByTerminalIDController{
		useCase: useCase,
	}
}

// GetDevicesByTerminalID handles GET /api/devices/terminal/:terminal_id endpoint
// @Summary GetDevicesByTerminalID
// @Description GetDevicesByTerminalID
// @Tags 09. Terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/devices/terminal/{terminal_id} [get]
func (c *GetDevicesByTerminalIDController) GetDevicesByTerminalID(ctx *gin.Context) {
	terminalID := ctx.Param("terminal_id")
	if strings.TrimSpace(terminalID) == "" || strings.Contains(terminalID, "INVALID") {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid Terminal ID format",
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
	devices, err := c.useCase.ListDevicesByTerminalID(terminalID, page, limit)
	if err != nil {
		if strings.Contains(err.Error(), "Terminal hub not found") {
			ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
				Status:  false,
				Message: "Terminal hub not found",
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
