package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"

	"github.com/gin-gonic/gin"
)

// UpdateTerminalAIEngineProfileController handles PUT /api/terminal/:id/ai-engine-profile
type UpdateTerminalAIEngineProfileController struct {
	useCase *usecases.UpdateTerminalAIEngineProfileUseCase
}

// NewUpdateTerminalAIEngineProfileController creates a new instance
func NewUpdateTerminalAIEngineProfileController(useCase *usecases.UpdateTerminalAIEngineProfileUseCase) *UpdateTerminalAIEngineProfileController {
	return &UpdateTerminalAIEngineProfileController{useCase: useCase}
}

// UpdateAIEngineProfile handles PUT /api/terminal/:id/ai-engine-profile
// @Summary      Update AI engine profile
// @Description  Sets or clears the AI engine profile (fast, standard) for a terminal. plaud is rejected as not yet available.
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        id       path   string  true  "Terminal ID"
// @Param        request  body   terminal_dtos.UpdateTerminalAIEngineProfileRequestDTO  true  "Profile update"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      422  {object}  dtos.ErrorResponse
// @Router       /api/terminal/{id}/ai-engine-profile [put]
// @Security     BearerAuth
func (c *UpdateTerminalAIEngineProfileController) UpdateAIEngineProfile(ctx *gin.Context) {
	id := ctx.Param("id")
	var req terminal_dtos.UpdateTerminalAIEngineProfileRequestDTO

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

	result, err := c.useCase.Update(id, &req)
	if err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		statusCode := http.StatusInternalServerError
		errorMessage := err.Error()
		if errorMessage == "Terminal not found" {
			statusCode = http.StatusNotFound
			errorMessage = "Terminal not found"
		}

		utils.LogError("UpdateTerminalAIEngineProfileController: %v", err)
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: errorMessage,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Updated successfully",
		Data:    result,
	})
}
