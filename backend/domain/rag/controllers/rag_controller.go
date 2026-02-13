package controllers

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"

	"github.com/gin-gonic/gin"
)

// RAGProcessor is an abstraction for RAG operations implemented by the usecase.
// This allows unit tests to provide a fake implementation.
type RAGProcessor interface {
	Control(text string, authToken string, onComplete func(string, *dtos.RAGStatusDTO)) (string, error)
	GetStatus(taskID string) (*dtos.RAGStatusDTO, error)
	Translate(text string, language string) (string, error)
	TranslateAsync(text string, language string) (string, error)
	Summary(text string, language string, context string, style string) (*dtos.RAGSummaryResponseDTO, error)
	SummaryAsync(text string, language string, context string, style string) (string, error)
}

type RAGController struct {
	usecase RAGProcessor
	config  *utils.Config
}

func NewRAGController(usecase RAGProcessor, cfg *utils.Config) *RAGController {
	return &RAGController{usecase: usecase, config: cfg}
}

// Control godoc
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
func (c *RAGController) Control(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	// Extract token from Bearer string
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 {
		authHeader = parts[1]
	}

	// Call UseCase via Control
	taskID, err := c.usecase.Control(req.Text, authHeader, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Failed to queue task", Details: err.Error()})
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

// Translate godoc
// @Summary Translate text to specified language
// @Description Translate text to a target language (default English) using the LLM. Best for short phrases/commands.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "Translation request"
// @Success 200 {object} dtos.StandardResponse{data=string}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
func (c *RAGController) Translate(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	translated, err := c.usecase.Translate(req.Text, req.Language)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Translation failed", Details: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dtos.StandardResponse{Status: true, Message: "Translation successful", Data: translated})
}

// TranslateAsync godoc
// @Summary Translate text to specified language (Async)
// @Description Translate text to a target language asynchronously. Returns a Task ID for polling.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "Translation request"
// @Success 202 {object} dtos.StandardResponse{data=map[string]string}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag/translate/async [post]
func (c *RAGController) TranslateAsync(ctx *gin.Context) {
	var req dtos.RAGRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	taskID, err := c.usecase.TranslateAsync(req.Text, req.Language)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Translation task failed to queue", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Translation task queued", Data: map[string]string{"task_id": taskID}})
}

// Summary godoc
// @Summary Generate meeting minutes summary
// @Description Transform a long transcription into professional meeting minutes. Supports target language (id/en), context, and style selection.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGSummaryRequestDTO true "Summary request"
// @Success 200 {object} dtos.StandardResponse{data=dtos.RAGSummaryResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag/summary [post]
func (c *RAGController) Summary(ctx *gin.Context) {
	var req dtos.RAGSummaryRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	result, err := c.usecase.Summary(req.Text, req.Language, req.Context, req.Style)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Summary generation failed", Details: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dtos.StandardResponse{Status: true, Message: "Summary generated successfully", Data: result})
}

// SummaryAsync godoc
// @Summary Generate meeting minutes summary (Async)
// @Description Generate meeting minutes summary asynchronously. Returns a Task ID for polling.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGSummaryRequestDTO true "Summary request"
// @Success 202 {object} dtos.StandardResponse{data=map[string]string}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag/summary/async [post]
func (c *RAGController) SummaryAsync(ctx *gin.Context) {
	var req dtos.RAGSummaryRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: "Invalid request", Details: err.Error()})
		return
	}

	taskID, err := c.usecase.SummaryAsync(req.Text, req.Language, req.Context, req.Style)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Summary task failed to queue", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{Status: true, Message: "Summary task queued", Data: map[string]string{"task_id": taskID}})
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
	if status.Status == "completed" && status.ExecutionResult != nil {
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
