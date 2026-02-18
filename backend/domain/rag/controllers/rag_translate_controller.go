package controllers

import (
	"net/http"
	"teralux_app/domain/common/utils"
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
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/rag/translate [post]
func (c *RAGTranslateController) Translate(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	taskID, err := c.translateUC.TranslateText(req.Text, req.Language)
	if err != nil {
		utils.LogError("RAGTranslateController.Translate: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Translation task queued", Data: map[string]string{"task_id": taskID}})
}
