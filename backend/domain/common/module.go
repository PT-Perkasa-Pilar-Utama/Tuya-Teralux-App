package common

import (
	"net/http"
	"sensio/domain/common/controllers"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/routes"
	"sensio/domain/common/services"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"

	"github.com/gin-gonic/gin"
)

// CommonModule encapsulates common domain components
type CommonModule struct {
	HealthController      *controllers.HealthController
	CacheController       *controllers.CacheController
	DocsController        *controllers.DocsController
	MqttService                  *infrastructure.MqttService
	DeviceInfoExternalController *controllers.DeviceInfoExternalController
	NotificationExternalController *controllers.NotificationExternalController
}

// NewCommonModule initializes the common domain components
func NewCommonModule(badger *infrastructure.BadgerService, vector *infrastructure.VectorService, mqttSvc *infrastructure.MqttService, terminalRepo terminal_repositories.ITerminalRepository) *CommonModule {
	bigSvc := services.NewDeviceInfoExternalService()
	
	// Initialize notification service
	notificationSvc := services.NewNotificationExternalService(terminalRepo, mqttSvc)

	return &CommonModule{
		HealthController:      controllers.NewHealthController(),
		CacheController:       controllers.NewCacheController(badger, vector),
		DocsController:               controllers.NewDocsController(),
		MqttService:                  mqttSvc,
		DeviceInfoExternalController: controllers.NewDeviceInfoExternalController(bigSvc),
		NotificationExternalController: controllers.NewNotificationExternalController(notificationSvc),
	}
}

// RegisterRoutes registers common routes
func (m *CommonModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Markdown Docs
	router.GET("/docs/*path", m.DocsController.ServeDocs)

	// OpenAPI 3.1 Routes (Primary docs endpoint)
	// Serve Swagger UI at /openapi
	router.StaticFS("/openapi", http.Dir("./docs/openapi"))

	// Protected Routes
	routes.SetupCacheRoutes(protected, m.CacheController)
	routes.SetupDeviceInfoExternalRoutes(protected, m.DeviceInfoExternalController)
	routes.SetupNotificationExternalRoutes(protected, m.NotificationExternalController)
}
