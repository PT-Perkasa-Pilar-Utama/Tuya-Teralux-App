package common

import (
	"teralux_app/docs/swagger"
	"teralux_app/domain/common/controllers"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/routes"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// CommonModule encapsulates common domain components
type CommonModule struct {
	HealthController *controllers.HealthController
	CacheController  *controllers.CacheController
	DocsController   *controllers.DocsController
	EmailController  *controllers.EmailController
	MqttService      *infrastructure.MqttService
}

// NewCommonModule initializes the common domain components
func NewCommonModule(badger *infrastructure.BadgerService, vector *infrastructure.VectorService, mqttSvc *infrastructure.MqttService) *CommonModule {
	return &CommonModule{
		HealthController: controllers.NewHealthController(),
		CacheController:  controllers.NewCacheController(badger, vector),
		DocsController:   controllers.NewDocsController(),
		EmailController:  controllers.NewEmailController(utils.GetConfig()),
		MqttService:      mqttSvc,
	}
}

// RegisterRoutes registers common routes
func (m *CommonModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Markdown Docs
	router.GET("/docs/*path", m.DocsController.ServeDocs)

	// Swagger Routes
	router.Static("/swagger-assets", "./docs/swagger-ui")
	router.GET("/swagger/*any", func(c *gin.Context) {
		// Dynamic Host and Scheme based on the request
		swagger.SwaggerInfo.Host = c.Request.Host

		// Handle proxies for scheme detection
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		swagger.SwaggerInfo.Schemes = []string{scheme}

		if c.Param("any") == "" || c.Param("any") == "/" || c.Param("any") == "/index.html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, swagger.CustomSwaggerHTML)
		} else {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		}
	})

	// Protected Routes
	routes.SetupCacheRoutes(protected, m.CacheController)
	routes.RegisterEmailRoutes(protected, m.EmailController)
}
