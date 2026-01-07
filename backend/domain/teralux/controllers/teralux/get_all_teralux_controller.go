package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
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
// @Summary      Get All Teralux
// @Description  Retrieves a list of all teralux devices
// @Tags         03. Teralux
// @Accept       json
// @Produce      json
// @Param        page      query  int     false  "Page number"
// @Param        limit     query  int     false  "Items per page"
// @Param        per_page  query  int     false  "Items per page (alias for limit)"
// @Param        room_id   query  string  false  "Filter by Room ID"
// @Success      200  {object}  dtos.StandardResponse{data=teralux_dtos.TeraluxListResponseDTO}
// @Failure      401  {object}  dtos.StandardResponse "Unauthorized"
// @Failure      500  {object}  dtos.StandardResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/teralux [get]
func (c *GetAllTeraluxController) GetAllTeralux(ctx *gin.Context) {
	var filter teralux_dtos.TeraluxFilterDTO
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		// If query params are invalid, we can just proceed with empty/default filter or return error.
		// Usually ignoring bad query params or defaulting is safe.
	}

	// Execute use case
	teraluxList, err := c.useCase.Execute(&filter)
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
