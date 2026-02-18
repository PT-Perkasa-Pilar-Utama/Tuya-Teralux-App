package routes

import (
	"teralux_app/domain/common/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterEmailRoutes(router *gin.RouterGroup, controller *controllers.EmailController) {
	emailGroup := router.Group("/api/email")
	emailGroup.POST("/send", controller.SendEmail)
}
