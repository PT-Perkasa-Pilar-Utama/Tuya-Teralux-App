package controllers

import (
	"net/http"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

// RAGTranslateController handles translation requests.
type RAGTranslateController struct {
	translateUC usecases.TranslateUseCase
}

func NewRAGTranslateController(translateUC usecases.TranslateUseCase) *RAGTranslateController {
	return &RAGTranslateController{
		translateUC: translateUC,
	}
}

// Translate handles POST /api/rag/translate
// @Summary Translate text to specified language
// @Description Translate text to a target language asynchronously. Returns a Task ID for polling.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "Translation request"
// @Success 202 {object} dtos.StandardResponse{data=map[string]string}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag/translate [post]
func (c *RAGTranslateController) Translate(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	taskID, err := c.translateUC.TranslateText(req.Text, req.Language)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Translation task failed to queue", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Translation task queued", Data: map[string]string{"task_id": taskID}})
}
