package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/device"

	"github.com/gin-gonic/gin"
)

// CreateDeviceController handles create device requests
type CreateDeviceController struct {
	useCase *usecases.CreateDeviceUseCase
}

// NewCreateDeviceController creates a new CreateDeviceController instance
func NewCreateDeviceController(useCase *usecases.CreateDeviceUseCase) *CreateDeviceController {
	return &CreateDeviceController{
		useCase: useCase,
	}
}

// CreateDevice handles POST /api/devices endpoint
// @Summary      Create Device
// @Description  Creates a new device under a teralux unit
// @Tags         04. Devices
// @Accept       json
// @Produce      json
// @Param        request  body      teralux_dtos.CreateDeviceRequestDTO  true  "Create Device Request"
// @Success      201      {object}  dtos.StandardResponse{data=teralux_dtos.CreateDeviceResponseDTO}  "New record created"
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.CreateDeviceResponseDTO}  "Idempotent: record updated"
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      422      {object}  dtos.StandardResponse "Validation Error"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/devices [post]
func (c *CreateDeviceController) CreateDevice(ctx *gin.Context) {
	var req teralux_dtos.CreateDeviceRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Data:    nil,
		})
		return
	}

	// Execute use case
	response, isNew, err := c.useCase.Execute(&req)
	if err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		// Handle specific errors like "Invalid teralux_id" as 422
		if err.Error() == "Invalid teralux_id: Teralux hub does not exist" {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid teralux_id: Teralux hub does not exist",
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

	statusCode := http.StatusCreated
	if !isNew {
		statusCode = http.StatusOK
	}

	ctx.JSON(statusCode, dtos.StandardResponse{
		Status:  true,
		Message: "Device created successfully",
		Data:    response,
	})
}
