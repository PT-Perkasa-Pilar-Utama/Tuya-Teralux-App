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
// @Success      201  {object}  dtos.StandardResponse
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      422  {object}  dtos.StandardResponse
// @Router       /api/terminal [post]
// @Security     BearerAuth
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

		statusCode := utils.GetErrorStatusCode(err)
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
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
