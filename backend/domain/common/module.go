package common

import (
	"net/http"
	"sensio/domain/common/controllers"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/routes"
	"sensio/domain/common/services"

	"github.com/gin-gonic/gin"
)

// CommonModule encapsulates common domain components
type CommonModule struct {
	HealthController      *controllers.HealthController
	CacheController       *controllers.CacheController
	DocsController        *controllers.DocsController
	MqttService           *infrastructure.MqttService
	BigExternalController *controllers.BigExternalController
}

// NewCommonModule initializes the common domain components
func NewCommonModule(badger *infrastructure.BadgerService, vector *infrastructure.VectorService, mqttSvc *infrastructure.MqttService) *CommonModule {
	bigSvc := services.NewBigExternalService()
	return &CommonModule{
		HealthController:      controllers.NewHealthController(),
		CacheController:       controllers.NewCacheController(badger, vector),
		DocsController:        controllers.NewDocsController(),
		MqttService:           mqttSvc,
		BigExternalController: controllers.NewBigExternalController(bigSvc),
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
	routes.SetupBigExternalRoutes(protected, m.BigExternalController)
}
