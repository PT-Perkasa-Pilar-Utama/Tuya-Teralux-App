package controllers

import (
	"net/http"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGChatController struct {
	chatUC usecases.ChatUseCase
}

func NewRAGChatController(chatUC usecases.ChatUseCase) *RAGChatController {
	return &RAGChatController{
		chatUC: chatUC,
	}
}

// Chat handles the AI Assistant chat/command classification.
// @Summary AI Assistant Chat
// @Description Classifies user prompt into Chat or Control and returns appropriate response or redirection.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGChatRequestDTO true "Chat Request"
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/rag/chat [post]
func (c *RAGChatController) Chat(ctx *gin.Context) {
	var req dtos.RAGChatRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	uid, _ := ctx.Get("uid")
	uidStr := ""
	if uid != nil {
		uidStr = uid.(string)
	}

	res, err := c.chatUC.Chat(uidStr, req.TeraluxID, req.Prompt, req.Language)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to process chat",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Chat processed successfully",
		Data:    res,
	})
}
