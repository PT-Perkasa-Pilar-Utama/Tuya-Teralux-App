package download_token

import (
	download_tokenControllers "sensio/domain/download_token/controllers"
	download_tokenRoutes "sensio/domain/download_token/routes"
	download_tokenServices "sensio/domain/download_token/services"

	"sensio/domain/infrastructure"

	"github.com/gin-gonic/gin"
)

type DownloadTokenModule struct {
	CreateTokenController  *download_tokenControllers.Handler
	ResolveTokenController *download_tokenControllers.Handler
}

func NewDownloadTokenModule(storageProvider infrastructure.StorageProvider) *DownloadTokenModule {
	service := download_tokenServices.NewDownloadTokenService(storageProvider)

	handler := download_tokenControllers.NewHandler(service)

	return &DownloadTokenModule{
		CreateTokenController:  handler,
		ResolveTokenController: handler,
	}
}

func (m *DownloadTokenModule) RegisterRoutes(router *gin.RouterGroup) {
	download_tokenRoutes.RegisterRoutes(router, m.CreateTokenController)
}

// Re-export for consumer convenience
type (
	DownloadTokenService = download_tokenServices.DownloadTokenService
)

// Re-export constructors
var (
	NewDownloadTokenService = download_tokenServices.NewDownloadTokenService
	NewHandler              = download_tokenControllers.NewHandler
	RegisterRoutes          = download_tokenRoutes.RegisterRoutes
)
