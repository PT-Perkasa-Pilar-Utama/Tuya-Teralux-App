package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"

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
// @Summary      Get all terminals
// @Description  Retrieve a list of all registered terminals with optional filtering
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        mac_address  query    string  false  "Filter by MAC address"
// @Param        name         query    string  false  "Filter by terminal name"
// @Param        room_id      query    string  false  "Filter by room ID"
// @Param        page         query    int     false  "Page number"
// @Param        limit        query    int     false  "Items per page"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.TerminalListResponseDTO}
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /api/terminal [get]
// @Security     BearerAuth
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
