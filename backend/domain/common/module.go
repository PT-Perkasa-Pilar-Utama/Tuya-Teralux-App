package common

import (
	"net/url"
	"teralux_app/docs"
	"teralux_app/domain/common/controllers"
	"teralux_app/domain/common/infrastructure/persistence"
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
}

// NewCommonModule initializes the common module
func NewCommonModule(badger *persistence.BadgerService) *CommonModule {
	return &CommonModule{
		HealthController: controllers.NewHealthController(),
		CacheController:  controllers.NewCacheController(badger),
	}
}

// RegisterRoutes registers common routes
func (m *CommonModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Configure Swagger Info
	if swaggerURL := utils.AppConfig.SwaggerBaseURL; swaggerURL != "" {
		if parsedURL, err := url.Parse(swaggerURL); err == nil {
			docs.SwaggerInfo.Host = parsedURL.Host
			docs.SwaggerInfo.Schemes = []string{parsedURL.Scheme}
		}
	}
	// Public Routes
	router.GET("/health", m.HealthController.CheckHealth)

	// Swagger Routes
	router.Static("/swagger-assets", "./docs/swagger-ui")
	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" || c.Param("any") == "/index.html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, docs.CustomSwaggerHTML)
		} else {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		}
	})

	// Protected Routes
	routes.SetupCacheRoutes(protected, m.CacheController)
}
