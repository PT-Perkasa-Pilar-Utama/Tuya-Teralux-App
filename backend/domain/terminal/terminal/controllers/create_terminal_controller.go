package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"

	"github.com/gin-gonic/gin"
)

// CreateTerminalController handles create terminal requests
type CreateTerminalController struct {
	useCase *usecases.CreateTerminalUseCase
}

// NewCreateTerminalController creates a new CreateTerminalController instance
func NewCreateTerminalController(useCase *usecases.CreateTerminalUseCase) *CreateTerminalController {
	return &CreateTerminalController{
		useCase: useCase,
	}
}

// CreateTerminal handles POST /api/terminal endpoint
// @Summary      Create a new terminal
// @Description  Register a new terminal device with MAC address and metadata
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        request  body      terminal_dtos.CreateTerminalRequestDTO  true  "Terminal registration data"
// @Success      201  {object}  dtos.StandardResponse{data=terminal_dtos.CreateTerminalResponseDTO}
// @Failure      422  {object}  dtos.ValidationErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/terminal [post]
func (c *CreateTerminalController) CreateTerminal(ctx *gin.Context) {
	var req terminal_dtos.CreateTerminalRequestDTO

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
	response, _, err := c.useCase.CreateTerminal(&req)
	if err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		// Check for specific error types and return appropriate status/message
		statusCode := utils.GetErrorStatusCode(err)
		errMsg := err.Error()

		// Map internal error messages to user-friendly messages
		var message string
		switch {
		case errMsg == "record not found" || errMsg == "Terminal not found":
			statusCode = http.StatusNotFound
			message = "Terminal not found"
		case errMsg == "Mac Address already in use" || errMsg == "MAC address already exists":
			statusCode = http.StatusConflict
			message = "MAC address already registered"
		case errMsg == "Room not found" || errMsg == "Invalid room_id":
			statusCode = http.StatusBadRequest
			message = "Invalid room ID"
		default:
			// Log internal error but return generic message to client
			utils.LogError("CreateTerminalController.CreateTerminal: %v", err)
			message = "Failed to create terminal"
		}

		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: message,
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Terminal created successfully",
		Data:    response,
	})
}
