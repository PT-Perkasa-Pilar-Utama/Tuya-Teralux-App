package routes

import (
	"sensio/domain/common/controllers"

	"github.com/gin-gonic/gin"
)

func SetupLoginRoutes(rg *gin.RouterGroup, controller *controllers.LoginController) {
	api := rg.Group("/api/common")
	{
		api.POST("/login", controller.Login)
	}
}