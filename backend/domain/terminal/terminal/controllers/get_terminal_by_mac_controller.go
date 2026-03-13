package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	usecases "sensio/domain/terminal/terminal/usecases"
	"strings"

	"github.com/gin-gonic/gin"
)

// Force usage for Swagger
var _ = terminal_dtos.TerminalSingleResponseDTO{}

// GetTerminalByMACController handles get terminal by MAC address requests
type GetTerminalByMACController struct {
	useCase *usecases.GetTerminalByMACUseCase
}

// NewGetTerminalByMACController creates a new GetTerminalByMACController instance
func NewGetTerminalByMACController(useCase *usecases.GetTerminalByMACUseCase) *GetTerminalByMACController {
	return &GetTerminalByMACController{
		useCase: useCase,
	}
}

// GetTerminalByMAC handles GET /api/terminal/mac/:mac endpoint
// @Summary      Get terminal by MAC address
// @Description  Retrieve terminal information by MAC address
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        mac  path  string  true  "Terminal MAC Address"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.TerminalSingleResponseDTO}
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Router       /api/terminal/mac/{mac} [get]
// @Security     BearerAuth
func (c *GetTerminalByMACController) GetTerminalByMAC(ctx *gin.Context) {
	mac := ctx.Param("mac")
	if strings.TrimSpace(mac) == "" || strings.Contains(mac, "INVALID") {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid MAC Address format",
			Data:    nil,
		})
		return
	}

	// Execute use case
	terminal, err := c.useCase.GetTerminalByMAC(mac)
	if err != nil {
		// Check for validation error
		if strings.Contains(strings.ToLower(err.Error()), "invalid mac") || strings.Contains(strings.ToLower(err.Error()), "format") {
			ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
				Status:  false,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}

		// Check if it's a "not found" error
		if err.Error() == "record not found" || strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
				Status:  false,
				Message: "Device not found",
				Data:    nil,
			})
			return
		}

		// For any other error, return generic internal server error
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal server error: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Terminal retrieved successfully",
		Data:    terminal,
	})
}
