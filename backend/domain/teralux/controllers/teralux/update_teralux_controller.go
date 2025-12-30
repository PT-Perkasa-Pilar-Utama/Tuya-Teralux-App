package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// UpdateTeraluxController handles update teralux requests
type UpdateTeraluxController struct {
	useCase *usecases.UpdateTeraluxUseCase
}

// NewUpdateTeraluxController creates a new UpdateTeraluxController instance
func NewUpdateTeraluxController(useCase *usecases.UpdateTeraluxUseCase) *UpdateTeraluxController {
	return &UpdateTeraluxController{
		useCase: useCase,
	}
}

// UpdateTeralux handles PUT /api/teralux/:id endpoint
// @Summary      Update Teralux
// @Description  Updates an existing teralux device
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Param        id       path      string                                true  "Teralux ID"
// @Param        request  body      teralux_dtos.UpdateTeraluxRequestDTO  true  "Update Teralux Request"
// @Success      200      {object}  dtos.StandardResponse
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      404      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/teralux/{id} [put]
func (c *UpdateTeraluxController) UpdateTeralux(ctx *gin.Context) {
	id := ctx.Param("id")
	var req teralux_dtos.UpdateTeraluxRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Execute use case
	err := c.useCase.Execute(id, &req)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to update teralux: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Teralux updated successfully",
		Data:    nil,
	})
}
