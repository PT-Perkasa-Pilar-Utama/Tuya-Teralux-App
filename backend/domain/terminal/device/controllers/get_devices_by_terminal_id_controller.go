package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/device/dtos"
	usecases "sensio/domain/terminal/device/usecases"
	"strconv"
	"strings"

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
// @Summary      Get devices by terminal ID
// @Description  Retrieve all devices associated with a specific terminal
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        terminal_id  path    string  true  "Terminal ID"
// @Param        page         query   int     false  "Page number"
// @Param        limit        query   int     false  "Items per page"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.DeviceListResponseDTO}
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Router       /api/devices/terminal/{terminal_id} [get]
// @Security     BearerAuth
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
