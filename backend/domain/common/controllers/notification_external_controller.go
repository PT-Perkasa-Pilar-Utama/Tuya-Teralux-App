package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/services"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"

	"github.com/gin-gonic/gin"
)

// NotificationExternalController handles HTTP requests for notifications
type NotificationExternalController struct {
	notificationSvc *services.NotificationExternalService
}

// NewNotificationExternalController creates a new instance of NotificationExternalController
func NewNotificationExternalController(notificationSvc *services.NotificationExternalService) *NotificationExternalController {
	return &NotificationExternalController{
		notificationSvc: notificationSvc,
	}
}

// PublishToRoom handles POST /api/notification/publish
// @Summary Publish a notification to all terminals in a room (with optional WhatsApp)
// @Description Publishes a notification to all terminals in the room via MQTT.
// @Description If phone_numbers is provided and device info is available, WhatsApp notifications are scheduled.
// @Description Requires room_id; phone_numbers optional. Optional scheduled_at (ISO 8601) for explicit time.
// @Description If scheduled_at is omitted, booking end time is derived from device info. Template defaults to end_meeting.
// @Tags 08. Common
// @Accept json
// @Produce json
// @Param        request body      terminal_dtos.NotificationPublishRequest true "Notification details (scheduled_at, template, and phone_numbers are optional)"
// @Success 200 {object} dtos.StandardResponse{data=terminal_dtos.NotificationPublishResponse}
// @Failure      400  {object}  dtos.ValidationErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router /api/notification/publish [post]
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
		// Check if it's an APIError (like 400 or 404)
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
