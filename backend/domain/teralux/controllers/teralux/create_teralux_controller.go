package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// CreateTeraluxController handles create teralux requests
type CreateTeraluxController struct {
	useCase *usecases.CreateTeraluxUseCase
}

// NewCreateTeraluxController creates a new CreateTeraluxController instance
func NewCreateTeraluxController(useCase *usecases.CreateTeraluxUseCase) *CreateTeraluxController {
	return &CreateTeraluxController{
		useCase: useCase,
	}
}

// CreateTeralux handles POST /api/teralux endpoint
// @Summary      Create Teralux
// @Description  Creates a new teralux device
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Param        request  body      teralux_dtos.CreateTeraluxRequestDTO  true  "Create Teralux Request"
// @Success      201      {object}  dtos.StandardResponse{data=teralux_dtos.CreateTeraluxResponseDTO}
// @Failure      400      {object}  dtos.StandardResponse
// @Failure      500      {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/teralux [post]
func (c *CreateTeraluxController) CreateTeralux(ctx *gin.Context) {
	var req teralux_dtos.CreateTeraluxRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Execute use case
	teralux, err := c.useCase.Execute(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to create teralux: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Teralux created successfully",
		Data:    teralux,
	})
}
