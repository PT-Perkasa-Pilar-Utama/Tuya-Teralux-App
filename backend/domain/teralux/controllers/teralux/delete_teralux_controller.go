package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	usecases "teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// DeleteTeraluxController handles delete teralux requests
type DeleteTeraluxController struct {
	useCase *usecases.DeleteTeraluxUseCase
}

// NewDeleteTeraluxController creates a new DeleteTeraluxController instance
func NewDeleteTeraluxController(useCase *usecases.DeleteTeraluxUseCase) *DeleteTeraluxController {
	return &DeleteTeraluxController{
		useCase: useCase,
	}
}

// DeleteTeralux handles DELETE /api/teralux/:id endpoint
// @Summary      Delete Teralux
// @Description  Soft deletes a teralux device by ID
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Teralux ID"
// @Success      200  {object}  dtos.StandardResponse "Successfully deleted"
// @Failure      400  {object}  dtos.StandardResponse "Invalid ID format"
// @Failure      401  {object}  dtos.StandardResponse "Unauthorized"
// @Failure      404  {object}  dtos.StandardResponse "Teralux not found"
// @Failure      500  {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/teralux/{id} [delete]
func (c *DeleteTeraluxController) DeleteTeralux(ctx *gin.Context) {
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
	if err := c.useCase.Execute(id); err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Teralux not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Teralux deleted successfully",
		Data:    nil,
	})
}
