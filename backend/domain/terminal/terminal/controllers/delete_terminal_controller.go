package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"
	"strings"

	"github.com/gin-gonic/gin"
)

// DeleteTerminalController handles delete terminal requests
type DeleteTerminalController struct {
	useCase *usecases.DeleteTerminalUseCase
}

// NewDeleteTerminalController creates a new DeleteTerminalController instance
func NewDeleteTerminalController(useCase *usecases.DeleteTerminalUseCase) *DeleteTerminalController {
	return &DeleteTerminalController{
		useCase: useCase,
	}
}

// DeleteTerminal handles DELETE /api/terminal/:id endpoint
// @Summary      Delete a terminal
// @Description  Soft delete a terminal by ID
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Terminal ID"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      400  {object}  dtos.ValidationErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Router       /api/terminal/{id} [delete]
// @Security     BearerAuth
func (c *DeleteTerminalController) DeleteTerminal(ctx *gin.Context) {
	id := ctx.Param("id")

	if strings.TrimSpace(id) == "" || strings.Contains(id, "INVALID") {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid ID format",
			Data:    nil,
		})
		return
	}

	// Execute use case
	if err := c.useCase.DeleteTerminal(id); err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Terminal not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Terminal deleted successfully",
		Data:    nil,
	})
}
