package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = terminal_dtos.TerminalSingleResponseDTO{}

// GetTerminalByIDController handles get terminal by ID requests
type GetTerminalByIDController struct {
	useCase *usecases.GetTerminalByIDUseCase
}

// NewGetTerminalByIDController creates a new GetTerminalByIDController instance
func NewGetTerminalByIDController(useCase *usecases.GetTerminalByIDUseCase) *GetTerminalByIDController {
	return &GetTerminalByIDController{
		useCase: useCase,
	}
}

// GetTerminalByID handles GET /api/terminal/:id endpoint
// @Summary      Get terminal by ID
// @Description  Retrieve terminal information by terminal ID
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Terminal ID"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.TerminalSingleResponseDTO}
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Router       /api/terminal/{id} [get]
// @Security     BearerAuth
func (c *GetTerminalByIDController) GetTerminalByID(ctx *gin.Context) {
	id := ctx.Param("id")

	// Execute use case (validation happens in use case)
	terminal, err := c.useCase.GetTerminalByID(id)
	if err != nil {
		// Check if it's a validation error
		if err.Error() == "Invalid ID format" {
			ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid ID format",
				Data:    nil,
			})
			return
		}

		// Otherwise it's not found
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Terminal not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Terminal retrieved successfully",
		Data:    terminal,
	})
}
