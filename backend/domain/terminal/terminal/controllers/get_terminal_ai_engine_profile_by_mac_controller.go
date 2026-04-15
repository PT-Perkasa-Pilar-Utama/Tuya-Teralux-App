package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	usecases "sensio/domain/terminal/terminal/usecases"

	"github.com/gin-gonic/gin"
)

// GetTerminalAIEngineProfileByMACController handles GET /api/terminal/mac/:mac/ai-engine-profile
type GetTerminalAIEngineProfileByMACController struct {
	useCase *usecases.GetTerminalAIEngineProfileUseCase
}

// NewGetTerminalAIEngineProfileByMACController creates a new instance
func NewGetTerminalAIEngineProfileByMACController(useCase *usecases.GetTerminalAIEngineProfileUseCase) *GetTerminalAIEngineProfileByMACController {
	return &GetTerminalAIEngineProfileByMACController{useCase: useCase}
}

// GetAIEngineProfile handles GET /api/terminal/mac/:mac/ai-engine-profile
// @Summary      Get AI engine profile by MAC
// @Description  Returns the AI engine profile (fast, standard) for a terminal identified by MAC address
// @Tags         02. Terminal
// @Produce      json
// @Param        mac  path  string  true  "Terminal MAC address"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Router       /api/terminal/mac/{mac}/ai-engine-profile [get]
// @Security     ApiKeyAuth
func (c *GetTerminalAIEngineProfileByMACController) GetAIEngineProfile(ctx *gin.Context) {
	mac := ctx.Param("mac")

	result, err := c.useCase.GetByMac(mac)
	if err != nil {
		errMsg := err.Error()
		statusCode := http.StatusInternalServerError
		if errMsg == "Terminal not found" {
			statusCode = http.StatusNotFound
		}
		utils.LogError("GetTerminalAIEngineProfileByMACController: %v", err)
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: http.StatusText(statusCode),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Success",
		Data:    result,
	})
}
