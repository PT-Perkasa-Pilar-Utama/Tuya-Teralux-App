package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/tasks"
	"sensio/domain/models/rag/dtos"

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

// GetStatus handles GET /api/models/rag/:task_id
// @Summary Get RAG task status
// @Tags 04. Models
// @Security BearerAuth
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} commonDtos.StandardResponse{data=dtos.RAGStatusDTO}
// @Failure      404  {object}  commonDtos.ErrorResponse
// @Router /api/models/rag/{task_id} [get]
func (c *RAGStatusController) GetStatus(ctx *gin.Context) {
	id := ctx.Param("task_id")
	status, err := c.statusUC.GetTaskStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, commonDtos.StandardResponse{
			Status:  false,
			Message: "Resource Not Found",
		})
		return
	}

	isSuccess := status.Status != "failed"
	message := "Task status retrieved successfully"
	httpStatus := http.StatusOK

	if status.Status == "failed" {
		message = "Task failed: " + status.Error
		if status.HTTPStatusCode != 0 {
			httpStatus = status.HTTPStatusCode
		}
	}

	ctx.JSON(httpStatus, commonDtos.StandardResponse{
		Status:  isSuccess,
		Message: message,
		Data:    status,
	})
}
