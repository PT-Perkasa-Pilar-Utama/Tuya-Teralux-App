package mail

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/mail/controllers"
	"teralux_app/domain/mail/routes"
	"teralux_app/domain/mail/services"
	"teralux_app/domain/mail/usecases"

	"github.com/gin-gonic/gin"
)

type MailModule struct {
	SendController      *controllers.MailSendController
	SendByMacController *controllers.MailSendByMacController
	UseCase             usecases.MailSendUseCase
	SendByMacUseCase    usecases.MailSendByMacUseCase
}

func NewMailModule(cfg *utils.Config) *MailModule {
	service := services.NewMailService(cfg)
	externalService := services.NewMailExternalService()

	useCase := usecases.NewMailSendUseCase(service)
	sendByMacUseCase := usecases.NewMailSendByMacUseCase(service, externalService)

	controller := controllers.NewMailSendController(useCase)
	sendByMacController := controllers.NewMailSendByMacController(sendByMacUseCase)

	return &MailModule{
		SendController:      controller,
		SendByMacController: sendByMacController,
		UseCase:             useCase,
		SendByMacUseCase:    sendByMacUseCase,
	}
}

func (m *MailModule) RegisterRoutes(router *gin.RouterGroup) {
	routes.RegisterMailSendRoutes(router, m.SendController, m.SendByMacController)
}
