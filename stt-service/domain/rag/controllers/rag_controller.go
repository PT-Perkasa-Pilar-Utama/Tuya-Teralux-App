package controllers

import (
	"net/http"
	"stt-service/domain/rag/dtos"
	"stt-service/domain/rag/usecases"
	speechDtos "stt-service/domain/speech/dtos"

	"github.com/gin-gonic/gin"
)

type RAGController struct {
	ragUsecase usecases.RAGUsecase
}

func NewRAGController(ragUsecase usecases.RAGUsecase) *RAGController {
	return &RAGController{
		ragUsecase: ragUsecase,
	}
}

// ProcessText handles the text input for RAG
// @Summary      Process text through RAG (mocked)
// @Description  Accepts text input and returns a simulated RAG process status
// @Tags         rag
// @Accept       json
// @Produce      json
// @Param        request body dtos.RAGRequest true "RAG Request"
// @Success      200 {object} speechDtos.StandardResponse{data=dtos.RAGResponse}
// @Router       /v1/rag [post]
func (c *RAGController) ProcessText(ctx *gin.Context) {
	var req dtos.RAGRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := c.ragUsecase.ProcessText(req.Text)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, speechDtos.StandardResponse{
		Status:  true,
		Message: "RAG process initiated",
		Data:    res,
	})
}

// GetStatus returns the status of a RAG task
// @Summary      Get RAG task status
// @Description  Check the status or get the result of a RAG task by ID
// @Tags         rag
// @Produce      json
// @Param        task_id path string true "Task ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200 {object} speechDtos.StandardResponse{data=dtos.RAGResponse}
// @Router       /v1/rag/{task_id} [get]
func (c *RAGController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	res, err := c.ragUsecase.GetStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, speechDtos.StandardResponse{
		Status:  true,
		Message: "RAG status retrieved",
		Data:    res,
	})
}
