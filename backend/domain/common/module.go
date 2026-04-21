package common

import (
	"github.com/gin-gonic/gin"
	"sensio/domain/common/controllers"
	"sensio/domain/common/routes"
	"sensio/domain/common/services"
	"sensio/domain/common/utils"
	"sensio/domain/download_token"
	"sensio/domain/infrastructure"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
)

// CommonModule encapsulates common domain components
type CommonModule struct {
	HealthController             *controllers.HealthController
	CacheController              *controllers.CacheController
	DocsController               *controllers.DocsController
	MqttService                  *infrastructure.MqttService
	DeviceInfoExternalController *controllers.DeviceInfoExternalController
	StorageProvider              infrastructure.StorageProvider
	DownloadTokenService         *download_token.DownloadTokenService
}

// NewCommonModule initializes the common domain components
func NewCommonModule(badger *infrastructure.BadgerService, vector *infrastructure.VectorService, mqttSvc *infrastructure.MqttService, terminalRepo terminal_repositories.ITerminalRepository, cfg *utils.Config) *CommonModule {
	// Initialize S3 storage provider
	storageProvider, err := infrastructure.NewStorageProvider(cfg)
	if err != nil {
		utils.LogWarn("CommonModule: Failed to initialize S3 storage: %v (using local fallback)", err)
		storageProvider, _ = infrastructure.NewStorageProvider(nil)
	}

	// Initialize download token service
	tokenService := download_token.NewDownloadTokenService(download_token.NewStore(), storageProvider)

	return &CommonModule{
		HealthController:             controllers.NewHealthController(),
		CacheController:              controllers.NewCacheController(badger, vector),
		DocsController:               controllers.NewDocsController(),
		MqttService:                  mqttSvc,
		DeviceInfoExternalController: controllers.NewDeviceInfoExternalController(services.NewDeviceInfoExternalService()),
		StorageProvider:              storageProvider,
		DownloadTokenService:         tokenService,
	}
}

// RegisterRoutes registers common routes
func (m *CommonModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Markdown Docs - render markdown with custom controller
	router.GET("/docs/*path", m.DocsController.ServeDocs)

	// OpenAPI 3.1 Routes with Scalar UI
	// Specific routes first, then catch-all
	router.GET("/openapi.json", func(c *gin.Context) {
		c.File("./docs/openapi/openapi.json")
	})
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("./docs/openapi/openapi.yaml")
	})
	router.GET("/openapi/", func(c *gin.Context) {
		c.File("./docs/openapi/index.html")
	})

	// Protected Routes
	routes.SetupCacheRoutes(protected, m.CacheController)
	routes.SetupDeviceInfoExternalRoutes(protected, m.DeviceInfoExternalController)

	// Download Token Routes
	downloadTokenHandler := download_token.NewHandler(m.DownloadTokenService)
	download_token.RegisterRoutes(protected, downloadTokenHandler)

	// Upload Intent Routes (removed - scope creep)
}
