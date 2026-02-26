package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/dtos"
	usecases "sensio/domain/terminal/usecases/device_status"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Force usage of terminal_dtos for Swagger
var _ = terminal_dtos.DeviceStatusListResponseDTO{}

// GetDeviceStatusesByDeviceIDController handles get device statuses by device ID requests
type GetDeviceStatusesByDeviceIDController struct {
	useCase *usecases.GetDeviceStatusesByDeviceIDUseCase
}

// NewGetDeviceStatusesByDeviceIDController creates a new GetDeviceStatusesByDeviceIDController instance
func NewGetDeviceStatusesByDeviceIDController(useCase *usecases.GetDeviceStatusesByDeviceIDUseCase) *GetDeviceStatusesByDeviceIDController {
	return &GetDeviceStatusesByDeviceIDController{
		useCase: useCase,
	}
}

// GetDeviceStatusesByDeviceID handles GET /api/device/statuses/:deviceId endpoint
func (c *GetDeviceStatusesByDeviceIDController) GetDeviceStatusesByDeviceID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "id", Message: "Device ID is required"},
			},
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

	statuses, err := c.useCase.ListDeviceStatusesByDeviceID(id, page, limit)
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
		Message: "Statuses retrieved successfully",
		Data:    statuses,
	})
}
