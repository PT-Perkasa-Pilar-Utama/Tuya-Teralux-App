package download_token

import (
	download_tokenControllers "sensio/domain/download_token/controllers"
	download_tokenRepositories "sensio/domain/download_token/repositories"
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
	store := download_tokenRepositories.NewStore()

	service := download_tokenServices.NewDownloadTokenService(store, storageProvider)

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
	Store                = download_tokenRepositories.Store
)

// Re-export constructors
var (
	NewDownloadTokenService = download_tokenServices.NewDownloadTokenService
	NewStore                = download_tokenRepositories.NewStore
	NewHandler              = download_tokenControllers.NewHandler
	RegisterRoutes          = download_tokenRoutes.RegisterRoutes
)