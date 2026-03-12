package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/usecases"

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

// Translate handles POST /api/models/rag/translate
// @Summary Translate text to specified language
// @Description Translate text to a target language asynchronously. Returns a Task ID for polling.
// @Tags 04. Models
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "Translation request"
// @Param Idempotency-Key header string false "Idempotency key to deduplicate requests"
// @Success 202 {object} commonDtos.StandardResponse{data=map[string]string}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/models/rag/translate [post]
func (c *RAGTranslateController) Translate(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	idempotencyKey := ctx.GetHeader("Idempotency-Key")

	taskID, err := c.translateUC.TranslateTextWithTrigger(req.Text, req.Language, ctx.Request.URL.Path, req.MacAddress, idempotencyKey)
	if err != nil {
		utils.LogError("RAGTranslateController.Translate: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{Status: true, Message: "Translation task queued", Data: map[string]string{"task_id": taskID}})
}
