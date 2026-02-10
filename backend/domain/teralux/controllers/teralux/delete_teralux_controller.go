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
