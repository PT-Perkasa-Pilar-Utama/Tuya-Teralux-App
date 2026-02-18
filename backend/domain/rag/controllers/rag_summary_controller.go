package controllers

import (
	"net/http"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

// RAGSummaryController handles summary requests.
type RAGSummaryController struct {
	summaryUC usecases.SummaryUseCase
}

func NewRAGSummaryController(summaryUC usecases.SummaryUseCase) *RAGSummaryController {
	return &RAGSummaryController{
		summaryUC: summaryUC,
	}
}

// Summary handles POST /api/rag/summary
// @Summary Generate meeting minutes summary
// @Description Generate meeting minutes summary asynchronously. Returns a Task ID for polling.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGSummaryRequestDTO true "Summary request"
// @Success 202 {object} dtos.StandardResponse{data=map[string]string}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/rag/summary [post]
func (c *RAGSummaryController) Summary(ctx *gin.Context) {
	var req dtos.RAGSummaryRequestDTO
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

	taskID, err := c.summaryUC.SummarizeText(req.Text, req.Language, req.Context, req.Style)
	if err != nil {
		utils.LogError("RAGSummaryController.Summary: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Summary task queued", Data: map[string]string{"task_id": taskID}})
}
