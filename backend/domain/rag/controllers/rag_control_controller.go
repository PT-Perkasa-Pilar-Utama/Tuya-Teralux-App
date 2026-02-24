package controllers

import (
	"net/http"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

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
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/rag/control [post]
func (c *RAGControlController) Control(ctx *gin.Context) {
	var req dtos.RAGControlRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
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

	res, err := c.controlUC.ProcessControl(uidStr, req.TeraluxID, req.Prompt)
	if err != nil {
		utils.LogError("RAGControlController.Control: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Control command processed successfully",
		Data:    res.Message,
	})
}
