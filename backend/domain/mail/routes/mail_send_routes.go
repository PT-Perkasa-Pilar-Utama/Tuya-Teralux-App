package routes

import (
	"teralux_app/domain/mail/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMailSendRoutes(router *gin.RouterGroup, sendController *controllers.MailSendController, sendByMacController *controllers.MailSendByMacController, statusController *controllers.MailStatusController) {
	mailGroup := router.Group("/api/mail")
	mailGroup.POST("/send", sendController.SendMail)
	mailGroup.POST("/send/mac/:mac_address", sendByMacController.SendMailByMac)
	mailGroup.GET("/status/:task_id", statusController.GetStatus)
}
