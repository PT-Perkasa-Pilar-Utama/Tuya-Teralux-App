package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGModelsOpenAIController interface {
	Query(ctx *gin.Context)
}

type ragModelsOpenAIController struct {
	usecase usecases.QueryOpenAIModelUseCase
}

// Force Swaggo to detect DTOs
var _ = dtos.RAGRawPromptRequestDTO{}

func NewRAGModelsOpenAIController(usecase usecases.QueryOpenAIModelUseCase) RAGModelsOpenAIController {
	return &ragModelsOpenAIController{usecase: usecase}
}

// Query godoc
// @Summary Raw prompt query to OpenAI model
// @Description Send a raw prompt directly to the OpenAI LLM model without RAG orchestration.
// @Tags 04. Models
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.RAGRawPromptRequestDTO true "Prompt Request"
// @Success 200 {object} commonDtos.StandardResponse{data=dtos.RAGRawPromptResponseDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      401  {object}  commonDtos.ErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router /api/models/openai [post]
func (c *ragModelsOpenAIController) Query(ctx *gin.Context) {
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
		utils.LogError("RAG OpenAI Raw Query failed: %v", err)
	}

	ctx.JSON(httpStatus, commonDtos.StandardResponse{
		Status:  isSuccess,
		Message: message,
		Data:    result,
	})
}
