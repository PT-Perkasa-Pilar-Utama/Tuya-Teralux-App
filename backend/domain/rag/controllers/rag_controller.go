package controllers

import (
	"net/http"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGController struct {
	usecase *usecases.RAGUsecase
}

func NewRAGController(u *usecases.RAGUsecase) *RAGController {
	return &RAGController{usecase: u}
}

// ProcessText godoc
// @Summary Process text via RAG
// @Description Submit text for RAG processing
// @Tags 09. RAG
// @Accept json
// @Produce json
// @Param request body dtos.RAGRequestDTO true "RAG request"
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /v1/rag [post]
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

	ctx.JSON(http.StatusOK, dtos.StandardResponse{Status: true, Message: "Task submitted", Data: map[string]string{"task_id": taskID}})
}

// GetStatus godoc
// @Summary Get RAG task status
// @Tags 09. RAG
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} dtos.StandardResponse
// @Failure 404 {object} dtos.StandardResponse
// @Router /v1/rag/{task_id} [get]
func (c *RAGController) GetStatus(ctx *gin.Context) {
	id := ctx.Param("task_id")
	status, err := c.usecase.GetStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{Status: false, Message: "Not found", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{Status: true, Message: "OK", Data: status})
}
