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
