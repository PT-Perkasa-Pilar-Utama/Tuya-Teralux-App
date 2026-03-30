package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGModelsLlamaCppController interface {
	Query(ctx *gin.Context)
}

type ragModelsLlamaCppController struct {
	usecase usecases.QueryLlamaCppModelUseCase
}

// Force Swaggo to detect DTOs
var _ = dtos.RAGRawPromptRequestDTO{}

func NewRAGModelsLlamaCppController(usecase usecases.QueryLlamaCppModelUseCase) RAGModelsLlamaCppController {
	return &ragModelsLlamaCppController{usecase: usecase}
}

// Query godoc
// @Summary Raw prompt query to local Llama.cpp model
// @Description Send a raw prompt directly to the local Llama.cpp LLM model without RAG orchestration.
// @Tags 04. Models
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.RAGRawPromptRequestDTO true "Prompt Request"
// @Success 200 {object} commonDtos.StandardResponse{data=dtos.RAGRawPromptResponseDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      401  {object}  commonDtos.ErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router /api/models/llama/cpp [post]
func (c *ragModelsLlamaCppController) Query(ctx *gin.Context) {
	var req dtos.RAGRawPromptRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Data:    err.Error(),
		})
		return
	}

	result, err := c.usecase.Query(ctx.Request.Context(), req.Prompt, ctx.Request.URL.Path)

	httpStatus := http.StatusOK
	message := "Query executed successfully"
	isSuccess := true

	if err != nil {
		isSuccess = false
		message = "Query execution failed"
		if result.HTTPStatusCode != 0 {
			httpStatus = result.HTTPStatusCode
		} else {
			httpStatus = http.StatusInternalServerError
		}
		utils.LogError("RAG Llama.cpp Raw Query failed: %v", err)
	}

	ctx.JSON(httpStatus, commonDtos.StandardResponse{
		Status:  isSuccess,
		Message: message,
		Data:    result,
	})
}
