package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/tasks"
	mail_dtos "teralux_app/domain/mail/dtos"

	"github.com/gin-gonic/gin"
)

// MailStatusController handles GET /api/email/status/:task_id.
type MailStatusController struct {
	statusUC tasks.GenericStatusUseCase[mail_dtos.MailStatusDTO]
}

func NewMailStatusController(statusUC tasks.GenericStatusUseCase[mail_dtos.MailStatusDTO]) *MailStatusController {
	return &MailStatusController{
		statusUC: statusUC,
	}
}

// GetStatus handles GET /api/email/status/:task_id
// @Summary Get email task status
// @Description Get the status and result of an email sending task.
// @Tags 09. Mail
// @Security BearerAuth
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} dtos.StandardResponse{data=mail_dtos.MailStatusDTO}
// @Failure 404 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/mail/status/{task_id} [get]
func (c *MailStatusController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Task ID is required",
		})
		return
	}

	status, err := c.statusUC.GetTaskStatus(taskID)
	if err == nil {
		isSuccess := status.Status != "failed"
		message := "Task status retrieved successfully"
		httpStatus := http.StatusOK

		if status.Status == "failed" {
			message = "Task failed"
			if status.HTTPStatusCode != 0 {
				httpStatus = status.HTTPStatusCode
			}
		}

		ctx.JSON(httpStatus, dtos.StandardResponse{
			Status:  isSuccess,
			Message: message,
			Data:    status,
		})
		return
	}

	ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
		Status:  false,
		Message: "Task not found",
	})
}
