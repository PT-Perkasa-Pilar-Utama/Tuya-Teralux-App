package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/models-v1/rag/services"

	"github.com/gin-gonic/gin"
)

type RAGController struct {
	ragSvc *services.PythonRAGService
}

func NewRAGController(ragSvc *services.PythonRAGService) *RAGController {
	return &RAGController{
		ragSvc: ragSvc,
	}
}

// Translate handles POST /api/models/v1/rag/translate
// @Summary      Translate text (v1)
// @Description  Translate text using the legacy Python RAG service
// @Tags         05. Models-v1
// @Accept       json
// @Produce      json
// @Param        request  body      services.RAGRequest  true  "Translation request"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/rag/translate [post]
// @Security     BearerAuth
func (c *RAGController) Translate(ctx *gin.Context) {
	var req services.RAGRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	resp, err := c.ragSvc.Translate(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}

// Summary handles POST /api/models/v1/rag/summary
// @Summary      Summarize text (v1)
// @Description  Summarize text using the legacy Python RAG service
// @Tags         05. Models-v1
// @Accept       json
// @Produce      json
// @Param        request  body      services.RAGRequest  true  "Summary request"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/rag/summary [post]
// @Security     BearerAuth
func (c *RAGController) Summary(ctx *gin.Context) {
	var req services.RAGRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	resp, err := c.ragSvc.Summary(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}

// Chat handles POST /api/models/v1/rag/chat
// @Summary      Chat with model (v1)
// @Description  Chat using the legacy Python RAG service
// @Tags         05. Models-v1
// @Accept       json
// @Produce      json
// @Param        request  body      services.RAGRequest  true  "Chat request"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/rag/chat [post]
// @Security     BearerAuth
func (c *RAGController) Chat(ctx *gin.Context) {
	var req services.RAGRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	resp, err := c.ragSvc.Chat(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}

// Control handles POST /api/models/v1/rag/control
// @Summary      Device control (v1)
// @Description  Control devices using the legacy Python RAG service
// @Tags         05. Models-v1
// @Accept       json
// @Produce      json
// @Param        request  body      services.RAGRequest  true  "Control request"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/rag/control [post]
// @Security     BearerAuth
func (c *RAGController) Control(ctx *gin.Context) {
	var req services.RAGRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	resp, err := c.ragSvc.Control(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}

// GetStatus handles GET /api/models/v1/rag/status/:task_id
// @Summary      Get RAG status (v1)
// @Description  Get the status of a RAG task by ID
// @Tags         05. Models-v1
// @Produce      json
// @Param        task_id  path      string  true  "Task ID"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/rag/status/{task_id} [get]
// @Security     BearerAuth
func (c *RAGController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	resp, err := c.ragSvc.GetStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}
