package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/teralux"

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
// @Success      201      {object}  dtos.StandardResponse{data=teralux_dtos.CreateTeraluxResponseDTO}  "New record created"
// @Success      200      {object}  dtos.StandardResponse{data=teralux_dtos.CreateTeraluxResponseDTO}  "Idempotent: record already exists"
// @Failure      401      {object}  dtos.StandardResponse "Unauthorized"
// @Failure      422      {object}  dtos.StandardResponse "Validation Error"
// @Failure      500      {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/teralux [post]
func (c *CreateTeraluxController) CreateTeralux(ctx *gin.Context) {
	var req teralux_dtos.CreateTeraluxRequestDTO

	// Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Data:    nil,
		})
		return
	}

	// Execute use case
	response, isNew, err := c.useCase.Execute(&req)
	if err != nil {
		if valErr, ok := err.(*utils.ValidationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
				Status:  false,
				Message: valErr.Message,
				Details: valErr.Details,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
			Data:    nil,
		})
		return
	}

	statusCode := http.StatusCreated
	if !isNew {
		statusCode = http.StatusOK
	}

	ctx.JSON(statusCode, dtos.StandardResponse{
		Status:  true,
		Message: "Teralux created successfully",
		Data:    response,
	})
}
