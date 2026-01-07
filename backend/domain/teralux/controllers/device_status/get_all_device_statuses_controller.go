package controllers

import (
	"net/http"
	"strconv"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device_status"

	"github.com/gin-gonic/gin"
)

// Force usage of teralux_dtos for Swagger
var _ = teralux_dtos.DeviceStatusListResponseDTO{}

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
// @Summary      Get All Device Statuses
// @Description  Retrieves all device statuses
// @Tags         05. Device Statuses
// @Accept       json
// @Produce      json
// @Param        page      query  int     false  "Page number"
// @Param        limit     query  int     false  "Items per page"
// @Param        per_page  query  int     false  "Items per page (alias for limit)"
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.DeviceStatusListResponseDTO}
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices/statuses [get]
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

	statuses, err := c.useCase.Execute(page, limit)
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
