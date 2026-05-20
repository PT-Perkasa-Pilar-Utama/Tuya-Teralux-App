package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sensio/domain/common/dtos"
	reports_dtos "sensio/domain/reports/dtos"
	"sensio/domain/reports/usecases"
)

var _ = reports_dtos.ReportResponseDto{}

type ReportsGetByIDController struct {
	useCase usecases.GetReportByIDUseCase
}

func NewReportsGetByIDController(useCase usecases.GetReportByIDUseCase) *ReportsGetByIDController {
	return &ReportsGetByIDController{
		useCase: useCase,
	}
}

func (c *ReportsGetByIDController) GetReportByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.useCase.GetReportByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Report not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Report retrieved successfully",
		Data:    result,
	})
}
