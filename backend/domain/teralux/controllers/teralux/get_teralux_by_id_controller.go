package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = teralux_dtos.TeraluxResponseDTO{}

// GetTeraluxByIDController handles get teralux by ID requests
type GetTeraluxByIDController struct {
	useCase *usecases.GetTeraluxByIDUseCase
}

// NewGetTeraluxByIDController creates a new GetTeraluxByIDController instance
func NewGetTeraluxByIDController(useCase *usecases.GetTeraluxByIDUseCase) *GetTeraluxByIDController {
	return &GetTeraluxByIDController{
		useCase: useCase,
	}
}

// GetTeraluxByID handles GET /api/teralux/:id endpoint
// @Summary      Get Teralux by ID
// @Description  Retrieves a single teralux device by ID with its associated devices
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Teralux ID"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.TeraluxResponseDTO}  "Returns teralux with room_id and devices array (empty if no devices)"
// @Failure      404  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/teralux/{id} [get]
func (c *GetTeraluxByIDController) GetTeraluxByID(ctx *gin.Context) {
	id := ctx.Param("id")

	// Execute use case
	teralux, err := c.useCase.Execute(id)
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
