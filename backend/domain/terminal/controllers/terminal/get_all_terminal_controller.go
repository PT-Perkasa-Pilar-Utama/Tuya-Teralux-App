package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/dtos"
	usecases "sensio/domain/terminal/usecases/terminal"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = terminal_dtos.TerminalListResponseDTO{}

// GetAllTerminalController handles get all terminal requests
type GetAllTerminalController struct {
	useCase *usecases.GetAllTerminalUseCase
}

// NewGetAllTerminalController creates a new GetAllTerminalController instance
func NewGetAllTerminalController(useCase *usecases.GetAllTerminalUseCase) *GetAllTerminalController {
	return &GetAllTerminalController{
		useCase: useCase,
	}
}

// GetAllTerminal handles GET /api/terminal endpoint
func (c *GetAllTerminalController) GetAllTerminal(ctx *gin.Context) {
	var filter terminal_dtos.TerminalFilterDTO
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		utils.LogWarn("GetAllTerminal: Failed to bind query filter: %v", err)
	}

	// Execute use case
	terminalList, err := c.useCase.ListTerminal(&filter)
	if err != nil {
		utils.LogError("GetAllTerminalController.GetAllTerminal: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Terminal list retrieved successfully",
		Data:    terminalList,
	})
}
