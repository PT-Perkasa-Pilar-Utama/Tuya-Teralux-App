package notification

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	notificationServices "sensio/domain/notification/services"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"

	"github.com/gin-gonic/gin"
)

type NotificationExternalController struct {
	notificationSvc *notificationServices.NotificationExternalService
}

func NewNotificationExternalController(notificationSvc *notificationServices.NotificationExternalService) *NotificationExternalController {
	return &NotificationExternalController{
		notificationSvc: notificationSvc,
	}
}

func (c *NotificationExternalController) PublishToRoom(ctx *gin.Context) {
	var req terminal_dtos.NotificationPublishRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.RoomID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "room_id is required",
		})
		return
	}

	if req.Template != "" && req.Template != "start_meeting" && req.Template != "end_meeting" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "template must be either 'start_meeting' or 'end_meeting'",
		})
		return
	}

	resp, err := c.notificationSvc.PublishNotificationToRoom(req)
	if err != nil {
		if apiErr, ok := err.(*utils.APIError); ok {
			ctx.JSON(apiErr.StatusCode, dtos.StandardResponse{
				Status:  false,
				Message: apiErr.Message,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to publish notification: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Notification published successfully",
		Data:    resp,
	})
}
