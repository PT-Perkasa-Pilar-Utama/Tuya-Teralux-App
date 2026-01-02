package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = teralux_dtos.TeraluxListResponseDTO{}

// GetAllTeraluxController handles get all teralux requests
type GetAllTeraluxController struct {
	useCase *usecases.GetAllTeraluxUseCase
}

// NewGetAllTeraluxController creates a new GetAllTeraluxController instance
func NewGetAllTeraluxController(useCase *usecases.GetAllTeraluxUseCase) *GetAllTeraluxController {
	return &GetAllTeraluxController{
		useCase: useCase,
	}
}

// GetAllTeralux handles GET /api/teralux endpoint
// @Summary      Get All Teralux
// @Description  Retrieves a list of all teralux devices
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.TeraluxListResponseDTO}
// @Failure      500  {object}  dtos.StandardResponse
// @Security     ApiKeyAuth
// @Router       /api/teralux [get]
func (c *GetAllTeraluxController) GetAllTeralux(ctx *gin.Context) {
	// Execute use case
	teraluxList, err := c.useCase.Execute()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to retrieve teralux list: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Teralux list retrieved successfully",
		Data:    teraluxList,
	})
}
