package common

import (
	"net/url"
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
}

// NewCommonModule initializes the common module
func NewCommonModule(badger *infrastructure.BadgerService) *CommonModule {
	return &CommonModule{
		HealthController: controllers.NewHealthController(),
		CacheController:  controllers.NewCacheController(badger),
		DocsController:   controllers.NewDocsController(),
	}
}

// RegisterRoutes registers common routes
func (m *CommonModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Configure Swagger Info
	if swaggerURL := utils.AppConfig.SwaggerBaseURL; swaggerURL != "" {
		if parsedURL, err := url.Parse(swaggerURL); err == nil {
			swagger.SwaggerInfo.Host = parsedURL.Host
			swagger.SwaggerInfo.Schemes = []string{parsedURL.Scheme}
		}
	}
	// Markdown Docs
	router.GET("/docs/*path", m.DocsController.ServeDocs)

	// Swagger Routes
	router.Static("/swagger-assets", "./docs/swagger-ui")
	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" || c.Param("any") == "/index.html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, swagger.CustomSwaggerHTML)
		} else {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		}
	})

	// Protected Routes
	routes.SetupCacheRoutes(protected, m.CacheController)
}
