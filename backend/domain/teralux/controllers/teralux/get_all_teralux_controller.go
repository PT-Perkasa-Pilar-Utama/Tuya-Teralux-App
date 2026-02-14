package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	teralux_dtos "teralux_app/domain/teralux/dtos"
	usecases "teralux_app/domain/teralux/usecases/teralux"

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
func (c *GetAllTeraluxController) GetAllTeralux(ctx *gin.Context) {
	var filter teralux_dtos.TeraluxFilterDTO
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		utils.LogWarn("GetAllTeralux: Failed to bind query filter: %v", err)
	}

	// Execute use case
	teraluxList, err := c.useCase.ListTeralux(&filter)
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
