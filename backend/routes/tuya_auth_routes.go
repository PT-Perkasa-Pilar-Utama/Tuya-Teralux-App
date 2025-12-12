package routes

import (
	"teralux_app/controllers"

	"github.com/gin-gonic/gin"
)

// SetupTuyaAuthRoutes registers Tuya authentication routes
func SetupTuyaAuthRoutes(router *gin.Engine, controller *controllers.TuyaAuthController) {
	api := router.Group("/api/tuya")
	{
		api.POST("/auth", controller.Authenticate)
	}
}
