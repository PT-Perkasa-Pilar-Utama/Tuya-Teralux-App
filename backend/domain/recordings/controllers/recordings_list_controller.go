package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/recordings/usecases"
)

type RecordingsListController struct {
	useCase usecases.GetAllRecordingsUseCase
}

func NewRecordingsListController(useCase usecases.GetAllRecordingsUseCase) *RecordingsListController {
	return &RecordingsListController{
		useCase: useCase,
	}
}

// ListRecordings handles GET /api/recordings endpoint
func (c *RecordingsListController) ListRecordings(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	result, err := c.useCase.ListRecordings(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
