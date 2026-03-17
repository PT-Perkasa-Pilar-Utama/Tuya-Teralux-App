package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/device_status/dtos"
	usecases "sensio/domain/terminal/device_status/usecases"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Force usage of terminal_dtos for Swagger
var _ = terminal_dtos.DeviceStatusListResponseDTO{}

// GetAllDeviceStatusesController handles get all device statuses requests
type GetAllDeviceStatusesController struct {
	useCase *usecases.GetAllDeviceStatusesUseCase
}

// NewGetAllDeviceStatusesController creates a new GetAllDeviceStatusesController instance
func NewGetAllDeviceStatusesController(useCase *usecases.GetAllDeviceStatusesUseCase) *GetAllDeviceStatusesController {
	return &GetAllDeviceStatusesController{
		useCase: useCase,
	}
}

// GetAllDeviceStatuses handles GET /api/devices/statuses endpoint
// @Summary      Get all device statuses
// @Description  Retrieve a list of all device statuses
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        page   query   int  false  "Page number"
// @Param        limit  query   int  false  "Items per page"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.DeviceStatusListResponseDTO}
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /api/devices/statuses [get]
// @Security     BearerAuth
func (c *GetAllDeviceStatusesController) GetAllDeviceStatuses(ctx *gin.Context) {
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

	statuses, err := c.useCase.ListDeviceStatuses(page, limit)
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
		Message: "Statuses retrieved successfully",
		Data:    statuses,
	})
}
