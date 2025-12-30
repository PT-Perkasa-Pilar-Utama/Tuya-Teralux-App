package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/teralux/usecases/teralux"

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
// @Success      200  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/teralux/{id} [delete]
func (c *DeleteTeraluxController) DeleteTeralux(ctx *gin.Context) {
	id := ctx.Param("id")

	// Execute use case
	err := c.useCase.Execute(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to delete teralux: " + err.Error(),
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
