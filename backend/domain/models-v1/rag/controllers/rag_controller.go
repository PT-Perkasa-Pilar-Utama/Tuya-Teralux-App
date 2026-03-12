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

func (c *RAGController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	resp, err := c.ragSvc.GetStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}
