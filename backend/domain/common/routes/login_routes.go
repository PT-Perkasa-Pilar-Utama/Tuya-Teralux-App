package routes

import (
	"sensio/domain/common/controllers"
	"sensio/domain/common/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupLoginRoutes(router *gin.Engine, controller *controllers.LoginController) {
	api := router.Group("/api/common")
	api.Use(middlewares.ApiKeyMiddleware())
	{
		api.POST("/login", controller.Login)
	}
}
