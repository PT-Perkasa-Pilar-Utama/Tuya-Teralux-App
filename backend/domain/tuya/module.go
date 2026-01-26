package tuya

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/middlewares"
	"teralux_app/domain/tuya/controllers"
	"teralux_app/domain/tuya/routes"
	"teralux_app/domain/tuya/services"
	"teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaModule encapsulates Tuya domain components
type TuyaModule struct {
	AuthController             *controllers.TuyaAuthController
	GetAllDevicesController    *controllers.TuyaGetAllDevicesController
	GetDeviceByIDController    *controllers.TuyaGetDeviceByIDController
	DeviceControlController    *controllers.TuyaDeviceControlController
	SensorController           *controllers.TuyaSensorController
	SyncDeviceStatusController *controllers.SyncDeviceStatusController

	// Exported Use Cases for other domains
	AuthUseCase          *usecases.TuyaAuthUseCase
	GetDeviceByIDUseCase *usecases.TuyaGetDeviceByIDUseCase
	DeviceControlUseCase *usecases.TuyaDeviceControlUseCase
}

// NewTuyaModule initializes the Tuya module
func NewTuyaModule(badger *infrastructure.BadgerService) *TuyaModule {
	// Services
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaDeviceService := services.NewTuyaDeviceService()

	// Use Cases
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)
	deviceStateUseCase := usecases.NewDeviceStateUseCase(badger)

	// Vector service for LLM-oriented indexing
	vectorSvc := infrastructure.NewVectorService()

	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService, deviceStateUseCase, badger, vectorSvc)
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaGetAllDevicesUseCase)
	tuyaDeviceControlUseCase := usecases.NewTuyaDeviceControlUseCase(tuyaDeviceService, deviceStateUseCase)
	tuyaSensorUseCase := usecases.NewTuyaSensorUseCase(tuyaGetDeviceByIDUseCase)
	syncDeviceStatusUseCase := usecases.NewSyncDeviceStatusUseCase(tuyaGetAllDevicesUseCase)

	// Controllers
	return &TuyaModule{
		AuthController:             controllers.NewTuyaAuthController(tuyaAuthUseCase),
		GetAllDevicesController:    controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase),
		GetDeviceByIDController:    controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase),
		DeviceControlController:    controllers.NewTuyaDeviceControlController(tuyaDeviceControlUseCase),
		SensorController:           controllers.NewTuyaSensorController(tuyaSensorUseCase),
		SyncDeviceStatusController: controllers.NewSyncDeviceStatusController(syncDeviceStatusUseCase),

		AuthUseCase:          tuyaAuthUseCase,
		GetDeviceByIDUseCase: tuyaGetDeviceByIDUseCase,
		DeviceControlUseCase: tuyaDeviceControlUseCase,
	}
}

// RegisterRoutes registers Tuya routes
func (m *TuyaModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Public/Auth Group
	authGroup := router.Group("/")
	authGroup.Use(middlewares.ApiKeyMiddleware())
	routes.SetupTuyaAuthRoutes(authGroup, m.AuthController)

	// Protected Routes
	routes.SetupTuyaDeviceRoutes(protected, m.GetAllDevicesController, m.GetDeviceByIDController, m.SensorController)
	routes.SetupTuyaControlRoutes(protected, m.DeviceControlController)

	// Sync Route
	protected.GET("/api/tuya/devices/sync", m.SyncDeviceStatusController.SyncStatus)
}
