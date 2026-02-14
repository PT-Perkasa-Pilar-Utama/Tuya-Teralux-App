package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

// RAGControlController handles control requests for RAG.
type RAGControlController struct {
	controlUC usecases.ControlUseCase
	statusUC  tasks.GenericStatusUseCase[dtos.RAGStatusDTO]
	config    *utils.Config
}

func NewRAGControlController(controlUC usecases.ControlUseCase, statusUC tasks.GenericStatusUseCase[dtos.RAGStatusDTO], cfg *utils.Config) *RAGControlController {
	return &RAGControlController{
		controlUC: controlUC,
		statusUC:  statusUC,
		config:    cfg,
	}
}

// Control handles POST /api/rag/control
// @Summary Control devices via natural language
// @Description Queue a RAG task to process natural language command
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "RAG Request"
// @Success 202 {object} dtos.StandardResponse{data=map[string]string}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag/control [post]
func (c *RAGControlController) Control(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 {
		authHeader = parts[1]
	}

	taskID, err := c.controlUC.ControlFromText(req.Text, authHeader, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Failed to queue task", Details: err.Error()})
		return
	}

	status, _ := c.statusUC.GetTaskStatus(taskID)
	if status != nil {
		ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]interface{}{"task_id": taskID, "task_status": status}})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]string{"task_id": taskID}})
}
