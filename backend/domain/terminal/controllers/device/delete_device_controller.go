package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	usecases "sensio/domain/terminal/usecases/device"

	"github.com/gin-gonic/gin"
)

// DeleteDeviceController handles delete device requests
type DeleteDeviceController struct {
	useCase *usecases.DeleteDeviceUseCase
}

// NewDeleteDeviceController creates a new DeleteDeviceController instance
func NewDeleteDeviceController(useCase *usecases.DeleteDeviceUseCase) *DeleteDeviceController {
	return &DeleteDeviceController{
		useCase: useCase,
	}
}

// DeleteDevice handles DELETE /api/devices/:id endpoint
func (c *DeleteDeviceController) DeleteDevice(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "id", Message: "ID is required"},
			},
		})
		return
	}

	if err := c.useCase.DeleteDevice(id); err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Device not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device deleted successfully",
		Data:    nil,
	})
}
