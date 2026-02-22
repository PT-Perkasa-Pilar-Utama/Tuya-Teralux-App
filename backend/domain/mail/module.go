package mail

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/mail/controllers"
	"teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/routes"
	"teralux_app/domain/mail/services"
	"teralux_app/domain/mail/usecases"

	"github.com/gin-gonic/gin"
)

type MailModule struct {
	SendController      *controllers.MailSendController
	SendByMacController *controllers.MailSendByMacController
	StatusController    *controllers.MailStatusController
	UseCase             usecases.MailSendUseCase
	SendByMacUseCase    usecases.MailSendByMacUseCase
}

func NewMailModule(cfg *utils.Config, badgerSvc *infrastructure.BadgerService) *MailModule {
	service := services.NewMailService(cfg)
	externalService := services.NewMailExternalService()

	// Task tracking
	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	cache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "mail:task:")
	statusUC := tasks.NewGenericStatusUseCase(cache, store)

	useCase := usecases.NewMailSendUseCase(service, store, cache)
	sendByMacUseCase := usecases.NewMailSendByMacUseCase(service, externalService, store, cache)

	controller := controllers.NewMailSendController(useCase)
	sendByMacController := controllers.NewMailSendByMacController(sendByMacUseCase)
	statusController := controllers.NewMailStatusController(statusUC)

	return &MailModule{
		SendController:      controller,
		SendByMacController: sendByMacController,
		StatusController:    statusController,
		UseCase:             useCase,
		SendByMacUseCase:    sendByMacUseCase,
	}
}

func (m *MailModule) RegisterRoutes(router *gin.RouterGroup) {
	routes.RegisterMailSendRoutes(router, m.SendController, m.SendByMacController, m.StatusController)
}
