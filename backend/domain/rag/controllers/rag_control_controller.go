package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

type RAGControlController struct {
	controlUC usecases.ControlUseCase
}

func NewRAGControlController(controlUC usecases.ControlUseCase) *RAGControlController {
	return &RAGControlController{
		controlUC: controlUC,
	}
}

// Control handles the redirected device control commands.
// @Summary AI Assistant Control
// @Description Processes natural language device control commands.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGControlRequestDTO true "Control Request"
// @Success 200 {object} commonDtos.StandardResponse
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/rag/control [post]
func (c *RAGControlController) Control(ctx *gin.Context) {
	var req dtos.RAGControlRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	uid, _ := ctx.Get("uid")
	uidStr := ""
	if uid != nil {
		uidStr = uid.(string)
	}

	res, err := c.controlUC.ProcessControl(uidStr, req.TerminalID, req.Prompt)
	if err != nil {
		utils.LogError("RAGControlController.Control: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Control command processed successfully",
		Data:    res.Message,
	})
}
