package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/usecases"
	"strings"

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

// Summary handles POST /api/models/rag/summary
// @Summary Generate meeting minutes summary
// @Description Generate meeting minutes summary asynchronously. Returns a Task ID for polling.
// @Tags 04. Models
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGSummaryRequestDTO true "Summary request"
// @Param Idempotency-Key header string false "Idempotency key to deduplicate requests"
// @Success 202 {object} commonDtos.StandardResponse{data=map[string]string}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router /api/models/rag/summary [post]
func (c *RAGSummaryController) Summary(ctx *gin.Context) {
	var req dtos.RAGSummaryRequestDTO
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

	participantsStr := strings.Join(req.Participants, ", ")
	idempotencyKey := ctx.GetHeader("Idempotency-Key")

	taskID, err := c.summaryUC.SummarizeTextWithTrigger(req.Text, req.Language, req.Context, req.Style, req.Date, req.Location, participantsStr, ctx.Request.URL.Path, req.MacAddress, idempotencyKey)
	if err != nil {
		utils.LogError("RAGSummaryController.Summary: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{Status: true, Message: "Summary task queued", Data: map[string]string{"task_id": taskID}})
}
