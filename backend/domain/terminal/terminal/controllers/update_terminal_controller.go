package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"
	"strings"

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
// @Summary      Update a terminal
// @Description  Update terminal information by ID
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        id       path    string                              true  "Terminal ID"
// @Param        request  body    terminal_dtos.UpdateTerminalRequestDTO  true  "Updated terminal data"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      422  {object}  dtos.StandardResponse
// @Router       /api/terminal/{id} [put]
// @Security     BearerAuth
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
		errMsg := err.Error()
		if errMsg == "Terminal not found" || errMsg == "record not found" || errMsg == "Terminal hub does not exist" {
			statusCode = http.StatusNotFound
		} else if strings.Contains(errMsg, "Mac Address already in use") {
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
