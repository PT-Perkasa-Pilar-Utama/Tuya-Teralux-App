package routes

import (
	"sensio/domain/common/controllers"

	"github.com/gin-gonic/gin"
)

func SetupLoginRoutes(router *gin.Engine, controller *controllers.LoginController) {
	api := router.Group("/api/common")
	{
		api.POST("/login", controller.Login)
	}
}