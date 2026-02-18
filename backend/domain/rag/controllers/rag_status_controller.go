package controllers

import (
	"net/http"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/rag/dtos"

	"github.com/gin-gonic/gin"
)

// RAGStatusController handles task status requests.
type RAGStatusController struct {
	statusUC tasks.GenericStatusUseCase[dtos.RAGStatusDTO]
}

func NewRAGStatusController(statusUC tasks.GenericStatusUseCase[dtos.RAGStatusDTO]) *RAGStatusController {
	return &RAGStatusController{
		statusUC: statusUC,
	}
}

// GetStatus handles GET /api/rag/:task_id
// @Summary Get RAG task status
// @Tags 05. RAG
// @Security BearerAuth
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} dtos.StandardResponse{data=dtos.RAGStatusDTO}
// @Failure 404 {object} dtos.StandardResponse
// @Router /api/rag/{task_id} [get]
func (c *RAGStatusController) GetStatus(ctx *gin.Context) {
	id := ctx.Param("task_id")
	status, err := c.statusUC.GetTaskStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Resource Not Found",
		})
		return
	}

	isSuccess := status.Status != "failed"
	message := "OK"
	if status.Status == "failed" {
		message = "Task failed"
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{Status: isSuccess, Message: message, Data: status})
}
