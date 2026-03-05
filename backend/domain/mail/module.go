package mail

import (
	"sensio/domain/common/infrastructure"
	commonServices "sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/mail/controllers"
	"sensio/domain/mail/dtos"
	"sensio/domain/mail/routes"
	"sensio/domain/mail/services"
	"sensio/domain/mail/usecases"

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
	externalService := commonServices.NewBigExternalService()

	// Task tracking
	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	cache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "cache:mail:task:")
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
