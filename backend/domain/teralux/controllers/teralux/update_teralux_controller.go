package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/teralux"

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
// @Success      200      {object}  dtos.StandardResponse "Successfully updated"
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      404      {object}  dtos.StandardResponse "Record not found"
// @Failure      409      {object}  dtos.StandardResponse "Conflict: Mac Address already in use"
// @Failure      422      {object}  dtos.StandardResponse "Validation Error"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/teralux/{id} [put]
func (c *UpdateTeraluxController) UpdateTeralux(ctx *gin.Context) {
	id := ctx.Param("id")
	var req teralux_dtos.UpdateTeraluxRequestDTO

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
	err := c.useCase.Execute(id, &req)
	if err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		// Check for specific error types/messages
		statusCode := http.StatusInternalServerError
		if err.Error() == "record not found" || err.Error() == "Teralux hub does not exist" {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Mac Address already in use") {
			statusCode = http.StatusConflict
		}

		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to update teralux: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Updated successfully",
		Data:    nil,
	})
}
