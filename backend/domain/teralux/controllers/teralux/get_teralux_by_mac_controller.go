package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// Force usage for Swagger
var _ = teralux_dtos.TeraluxSingleResponseDTO{}

// GetTeraluxByMACController handles get teralux by MAC address requests
type GetTeraluxByMACController struct {
	useCase *usecases.GetTeraluxByMACUseCase
}

// NewGetTeraluxByMACController creates a new GetTeraluxByMACController instance
func NewGetTeraluxByMACController(useCase *usecases.GetTeraluxByMACUseCase) *GetTeraluxByMACController {
	return &GetTeraluxByMACController{
		useCase: useCase,
	}
}

// GetTeraluxByMAC handles GET /api/teralux/mac/:mac endpoint
func (c *GetTeraluxByMACController) GetTeraluxByMAC(ctx *gin.Context) {
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
	teralux, err := c.useCase.GetTeraluxByMAC(mac)
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
		Message: "Teralux retrieved successfully",
		Data:    teralux,
	})
}
