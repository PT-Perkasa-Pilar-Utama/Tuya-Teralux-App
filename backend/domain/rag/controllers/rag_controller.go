package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/rag/dtos"

	"github.com/gin-gonic/gin"
)

// RAGProcessor is an abstraction for RAG operations implemented by the usecase.
// This allows unit tests to provide a fake implementation.
type RAGProcessor interface {
	Process(text string, authToken string) (string, error)
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
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "RAG request"
// @Success 202 {object} dtos.StandardResponse{data=dtos.RAGProcessResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag [post]
func (c *RAGController) ProcessText(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	authToken := ctx.GetHeader("Authorization")

	taskID, err := c.usecase.Process(req.Text, authToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Processing failed", Details: err.Error()})
		return
	}

	// Try to fetch cached status via DTO to include TTL info in response (optional)
	status, _ := c.usecase.GetStatus(taskID)
	// Use DTOs directly in the response, avoid hardcoding TTL values here
	if status != nil {
		ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]interface{}{"task_id": taskID, "task_status": status}})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]string{"task_id": taskID}})
}

// GetStatus godoc
// @Summary Get RAG task status
// @Tags 05. RAG
// @Security BearerAuth
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} dtos.StandardResponse{data=dtos.RAGStatusDTO}
// @Failure 404 {object} dtos.StandardResponse
// @Router /api/rag/{task_id} [get]
func (c *RAGController) GetStatus(ctx *gin.Context) {
	id := ctx.Param("task_id")
	status, err := c.usecase.GetStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{Status: false, Message: "Not found", Details: err.Error()})
		return
	}

	httpStatus := c.determineHTTPStatus(status)
	message := "OK"
	if httpStatus >= 400 {
		message = http.StatusText(httpStatus)
	}
	ctx.JSON(httpStatus, dtos.StandardResponse{Status: httpStatus < 400, Message: message, Data: status})
}

func (c *RAGController) determineHTTPStatus(status *dtos.RAGStatusDTO) int {
	if status.Status == "error" {
		if containsAny(status.Result, "429", "quota", "RESOURCE_EXHAUSTED") {
			return http.StatusTooManyRequests
		}
		if containsAny(status.Result, "401", "Unauthorized", "token expired", "Token expired") {
			return http.StatusUnauthorized
		}
		return http.StatusInternalServerError
	}

	// Also check individual execution result if task is done but failed internally
	if status.Status == "done" && status.ExecutionResult != nil {
		// Logic to check if execution result indicates an error
		if res, ok := status.ExecutionResult.(map[string]interface{}); ok {
			if s, ok := res["status"].(bool); ok && !s {
				msg, _ := res["message"].(string)
				if containsAny(msg, "Token expired", "Unauthorized") {
					return http.StatusUnauthorized
				}
				// For other execution errors, we might still want 200 OK
				// as the RAG task technically succeeded in making a decision,
				// but let's be strict if the user wants "error"
				return http.StatusBadRequest
			}
		}
	}

	return http.StatusOK
}

func containsAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(strings.ToLower(s), strings.ToLower(k)) {
			return true
		}
	}
	return false
}
