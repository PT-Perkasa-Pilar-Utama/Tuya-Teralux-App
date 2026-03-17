package routes

import (
	"sensio/domain/common/controllers"

	"github.com/gin-gonic/gin"
)

// SetupNotificationExternalRoutes registers notification proxy endpoints.
//
// param rg The router group to attach the routes to.
// param controller The controller handling notification operations.
func SetupNotificationExternalRoutes(rg *gin.RouterGroup, controller *controllers.NotificationExternalController) {
	notificationGroup := rg.Group("/api/notification")
	{
		// POST /api/notification/publish
		// Publishes computed notification time to room terminals
		notificationGroup.POST("/publish", controller.PublishToRoom)
	}
}
