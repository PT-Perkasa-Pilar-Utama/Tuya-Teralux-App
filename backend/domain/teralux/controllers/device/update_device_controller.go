package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// UpdateDeviceController handles update device requests
type UpdateDeviceController struct {
	useCase *usecases.UpdateDeviceUseCase
}

// NewUpdateDeviceController creates a new UpdateDeviceController instance
func NewUpdateDeviceController(useCase *usecases.UpdateDeviceUseCase) *UpdateDeviceController {
	return &UpdateDeviceController{
		useCase: useCase,
	}
}

// UpdateDevice handles PUT /api/devices/:id endpoint
// @Summary      Update Device
// @Description  Updates an existing device
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Device ID"
// @Param        request  body      teralux_dtos.UpdateDeviceRequestDTO  true  "Update Device Request"
// @Success      200      {object}  dtos.StandardResponse "Successfully updated"
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      404      {object}  dtos.StandardResponse "Device not found"
// @Failure      422      {object}  dtos.StandardResponse "Validation Error"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices/{id} [put]
func (c *UpdateDeviceController) UpdateDevice(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{{Field: "id", Message: "Device ID is required"}},
		})
		return
	}

	var req teralux_dtos.UpdateDeviceRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
		})
		return
	}

	if err := c.useCase.Execute(id, &req); err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		// Handle Not Found
		statusCode := http.StatusInternalServerError
		errorMsg := "Internal Server Error"
		if err.Error() == "record not found" || err.Error() == "Device not found" {
			statusCode = http.StatusNotFound
			errorMsg = "Device not found"
		}

		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: errorMsg,
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device updated successfully",
		Data:    nil,
	})
}
