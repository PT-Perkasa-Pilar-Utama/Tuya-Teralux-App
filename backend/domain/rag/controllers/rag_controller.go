package controllers

import (
	"net/http"
	"teralux_app/domain/rag/dtos"
	"github.com/gin-gonic/gin"
)

// RAGProcessor is an abstraction for RAG operations implemented by the usecase.
// This allows unit tests to provide a fake implementation.
type RAGProcessor interface {
	Process(text string) (string, error)
	GetStatus(taskID string) (*dtos.RAGStatusDTO, error)
}

type RAGController struct {
	usecase RAGProcessor
}

func NewRAGController(u RAGProcessor) *RAGController {
	return &RAGController{usecase: u}
}

// ProcessText godoc
// @Summary Process text via RAG
// @Description Submit text for RAG processing
// @Tags 09. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "RAG request"
// @Success 202 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag [post]
func (c *RAGController) ProcessText(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	taskID, err := c.usecase.Process(req.Text)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Processing failed", Details: err.Error()})
		return
	}

	// Try to fetch cached status via DTO to include TTL info in response (optional)
	status, _ := c.usecase.GetStatus(taskID)
	// Use DTOs directly in the response, avoid hardcoding TTL values here
	if status != nil {
		ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]interface{}{"task_id": taskID, "status": status}})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]string{"task_id": taskID}})
}

// GetStatus godoc
// @Summary Get RAG task status
// @Tags 09. RAG
// @Security BearerAuth
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} dtos.StandardResponse
// @Failure 404 {object} dtos.StandardResponse
// @Router /api/rag/{task_id} [get]
func (c *RAGController) GetStatus(ctx *gin.Context) {
	id := ctx.Param("task_id")
	status, err := c.usecase.GetStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{Status: false, Message: "Not found", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{Status: true, Message: "OK", Data: status})
}
