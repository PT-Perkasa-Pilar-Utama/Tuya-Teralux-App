package controllers

import (
	"net/http"
	"strings"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/dtos"
	usecases "sensio/domain/terminal/usecases/terminal"

	"github.com/gin-gonic/gin"
)

// UpdateTerminalController handles update terminal requests
type UpdateTerminalController struct {
	useCase *usecases.UpdateTerminalUseCase
}

// NewUpdateTerminalController creates a new UpdateTerminalController instance
func NewUpdateTerminalController(useCase *usecases.UpdateTerminalUseCase) *UpdateTerminalController {
	return &UpdateTerminalController{
		useCase: useCase,
	}
}

// UpdateTerminal handles PUT /api/terminal/:id endpoint
// @Summary UpdateTerminal
// @Description UpdateTerminal
// @Tags 09. Terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/terminal/{id} [put]
func (c *UpdateTerminalController) UpdateTerminal(ctx *gin.Context) {
	id := ctx.Param("id")
	var req terminal_dtos.UpdateTerminalRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	// Execute use case
	err := c.useCase.UpdateTerminal(id, &req)
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
		if err.Error() == "record not found" || err.Error() == "Terminal hub does not exist" {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Mac Address already in use") {
			statusCode = http.StatusConflict
		}

		utils.LogError("UpdateTerminalController.UpdateTerminal: %v", err)
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: http.StatusText(statusCode),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Updated successfully",
		Data:    nil,
	})
}
