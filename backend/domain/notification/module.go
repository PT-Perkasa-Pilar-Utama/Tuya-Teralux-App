package notification

import (
	"sensio/domain/common/services"
	"sensio/domain/infrastructure"
	notificationControllers "sensio/domain/notification/controllers"
	notificationRepositories "sensio/domain/notification/repositories"
	notificationRoutes "sensio/domain/notification/routes"
	notificationServices "sensio/domain/notification/services"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"

	"github.com/gin-gonic/gin"
)

type NotificationModule struct {
	NotificationExternalController *notificationControllers.NotificationExternalController
}

func NewNotificationModule(
	badger *infrastructure.BadgerService,
	mqttSvc infrastructure.IMqttService,
	terminalRepo terminal_repositories.ITerminalRepository,
) *NotificationModule {
	scheduledRepo := notificationRepositories.NewScheduledNotificationRepository()
	deviceInfoSvc := services.NewDeviceInfoExternalService()

	notificationExternalService := notificationServices.NewNotificationExternalServiceWithWA(
		terminalRepo,
		scheduledRepo,
		deviceInfoSvc,
		mqttSvc,
	)

	notificationExternalController := notificationControllers.NewNotificationExternalController(notificationExternalService)

	return &NotificationModule{
		NotificationExternalController: notificationExternalController,
	}
}

func (m *NotificationModule) RegisterRoutes(protected *gin.RouterGroup) {
	notificationRoutes.SetupNotificationExternalRoutes(protected, m.NotificationExternalController)
}
