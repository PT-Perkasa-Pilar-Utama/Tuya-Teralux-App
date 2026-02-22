package controllers

import (
	"net/http"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGModelsLlamaCppController interface {
	Query(ctx *gin.Context)
}

type ragModelsLlamaCppController struct {
	usecase usecases.QueryLlamaCppModelUseCase
}

func NewRAGModelsLlamaCppController(usecase usecases.QueryLlamaCppModelUseCase) RAGModelsLlamaCppController {
	return &ragModelsLlamaCppController{usecase: usecase}
}

// Query godoc
// @Summary Raw prompt query to local Llama.cpp model
// @Description Send a raw prompt directly to the local Llama.cpp LLM model without RAG orchestration.
// @Tags 06. Models
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.RAGRawPromptRequestDTO true "Prompt Request"
// @Success 200 {object} dtos.StandardResponse{data=dtos.RAGRawPromptResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 401 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/models/llama/cpp [post]
func (c *ragModelsLlamaCppController) Query(ctx *gin.Context) {
	var req dtos.RAGRawPromptRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Data:    err.Error(),
		})
		return
	}

	result, err := c.usecase.Query(req.Prompt, ctx.Request.URL.Path)
	
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

	ctx.JSON(httpStatus, dtos.StandardResponse{
		Status:  isSuccess,
		Message: message,
		Data:    result,
	})
}
