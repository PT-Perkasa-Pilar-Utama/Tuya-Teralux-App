package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGModelsGeminiController interface {
	Query(ctx *gin.Context)
}

type ragModelsGeminiController struct {
	usecase usecases.QueryGeminiModelUseCase
}

func NewRAGModelsGeminiController(usecase usecases.QueryGeminiModelUseCase) RAGModelsGeminiController {
	return &ragModelsGeminiController{usecase: usecase}
}

// Query godoc
// @Summary Raw prompt query to Gemini model
// @Description Send a raw prompt directly to the Gemini LLM model without RAG orchestration.
// @Tags 06. Models
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.RAGRawPromptRequestDTO true "Prompt Request"
// @Success 200 {object} commonDtos.StandardResponse{data=dtos.RAGRawPromptResponseDTO}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 401 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse
// @Router /api/models/gemini [post]
func (c *ragModelsGeminiController) Query(ctx *gin.Context) {
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
		utils.LogError("RAG Gemini Raw Query failed: %v", err)
	}

	ctx.JSON(httpStatus, commonDtos.StandardResponse{
		Status:  isSuccess,
		Message: message,
		Data:    result,
	})
}
