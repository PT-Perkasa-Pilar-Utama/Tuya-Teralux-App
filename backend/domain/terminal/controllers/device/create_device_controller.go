package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/dtos"
	usecases "sensio/domain/terminal/usecases/device"

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
func (c *CreateDeviceController) CreateDevice(ctx *gin.Context) {
	var req terminal_dtos.CreateDeviceRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	// Execute use case
	response, _, err := c.useCase.CreateDevice(&req)
	if err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		// Handle specific errors like "Invalid terminal_id" as 422
		if err.Error() == "Invalid terminal_id: Terminal hub does not exist" {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: "Validation Error",
				Details: []utils.ValidationErrorDetail{
					{Field: "terminal_id", Message: "Terminal hub does not exist"},
				},
			})
			return
		}

		// Log the actual error for debugging
		utils.LogError("CreateDeviceController: Internal Server Error: %v", err)

		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
			Data:    nil,
		})
		return
	}

	// Device created successfully (isNew should always be true now)
	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Device created successfully",
		Data:    response,
	})
}
