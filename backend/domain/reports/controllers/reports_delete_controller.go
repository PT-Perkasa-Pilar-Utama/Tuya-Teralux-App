package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/reports/usecases"
)

var _ = dtos.StandardResponse{}

type ReportsDeleteController struct {
	useCase usecases.DeleteReportUseCase
}

func NewReportsDeleteController(useCase usecases.DeleteReportUseCase) *ReportsDeleteController {
	return &ReportsDeleteController{
		useCase: useCase,
	}
}

func (c *ReportsDeleteController) DeleteReport(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.useCase.DeleteReport(id)
	if err != nil {
		utils.LogError("ReportsDeleteController.DeleteReport: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Report deleted successfully",
	})
}
