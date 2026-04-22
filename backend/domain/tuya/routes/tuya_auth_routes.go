package routes

import (
	"sensio/domain/common/middlewares"
	"sensio/domain/common/utils"
	"sensio/domain/tuya/controllers"

	"github.com/gin-gonic/gin"
)

// SetupTuyaAuthRoutes registers authentication-related endpoints for Tuya.
func SetupTuyaAuthRoutes(router *gin.Engine, controller *controllers.TuyaAuthController) {
	utils.LogDebug("SetupTuyaAuthRoutes initialized")
	authGroup := router.Group("/")
	authGroup.Use(middlewares.ApiKeyMiddleware())
	api := authGroup.Group("/api/tuya")
	api.GET("/auth", controller.Authenticate)
}
