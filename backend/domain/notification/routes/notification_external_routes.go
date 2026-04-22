package notification

import (
	notificationControllers "sensio/domain/notification/controllers"

	"github.com/gin-gonic/gin"
)

func SetupNotificationExternalRoutes(rg *gin.RouterGroup, controller *notificationControllers.NotificationExternalController) {
	notificationGroup := rg.Group("/api/notification")
	{
		notificationGroup.POST("/publish", controller.PublishToRoom)
	}
}
