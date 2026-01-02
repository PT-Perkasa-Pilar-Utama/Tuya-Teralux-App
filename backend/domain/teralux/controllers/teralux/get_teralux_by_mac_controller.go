package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

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
// @Summary      Get Teralux by MAC Address
// @Description  Retrieves a teralux device by its MAC address
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Param        mac  path      string  true  "MAC Address"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.TeraluxResponseDTO}
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     ApiKeyAuth
// @Router       /api/teralux/mac/{mac} [get]
func (c *GetTeraluxByMACController) GetTeraluxByMAC(ctx *gin.Context) {
	mac := ctx.Param("mac")

	// Execute use case
	teralux, err := c.useCase.Execute(mac)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Teralux not found: " + err.Error(),
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
