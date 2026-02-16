package tuya

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/middlewares"
	teralux_repositories "teralux_app/domain/teralux/repositories"
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
	CommandSwitchController    *controllers.TuyaCommandSwitchController
	SendIRCommandController    *controllers.TuyaSendIRCommandController
	SensorController           *controllers.TuyaSensorController
	SyncDeviceStatusController *controllers.SyncDeviceStatusController

	// Exported Use Cases for other domains
	AuthUseCase          usecases.TuyaAuthUseCase
	GetAllDevicesUseCase usecases.TuyaGetAllDevicesUseCase
	GetDeviceByIDUseCase *usecases.TuyaGetDeviceByIDUseCase
	DeviceControlUseCase usecases.TuyaDeviceControlExecutor
}

// NewTuyaModule initializes the Tuya module
func NewTuyaModule(badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, deviceRepo *teralux_repositories.DeviceRepository, teraluxRepo *teralux_repositories.TeraluxRepository) *TuyaModule {
	// Services
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaDeviceService := services.NewTuyaDeviceService()

	// Use Cases
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)
	deviceStateUseCase := usecases.NewDeviceStateUseCase(badger)

	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService, deviceStateUseCase, badger, vectorSvc, deviceRepo, teraluxRepo)
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService, badger)
	tuyaCommandSwitchUseCase := usecases.NewTuyaCommandSwitchUseCase(tuyaDeviceService, deviceStateUseCase)
	tuyaSendIRCommandUseCase := usecases.NewTuyaSendIRCommandUseCase(tuyaDeviceService)

	// Bridge for shared executor
	tuyaDeviceControlBridge := usecases.NewTuyaDeviceControlBridge(tuyaCommandSwitchUseCase, tuyaSendIRCommandUseCase)

	tuyaSensorUseCase := usecases.NewTuyaSensorUseCase(tuyaGetDeviceByIDUseCase)
	syncDeviceStatusUseCase := usecases.NewSyncDeviceStatusUseCase(tuyaGetAllDevicesUseCase)

	// Controllers
	return &TuyaModule{
		AuthController:             controllers.NewTuyaAuthController(tuyaAuthUseCase),
		GetAllDevicesController:    controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase),
		GetDeviceByIDController:    controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase),
		CommandSwitchController:    controllers.NewTuyaCommandSwitchController(tuyaCommandSwitchUseCase),
		SendIRCommandController:    controllers.NewTuyaSendIRCommandController(tuyaSendIRCommandUseCase),
		SensorController:           controllers.NewTuyaSensorController(tuyaSensorUseCase),
		SyncDeviceStatusController: controllers.NewSyncDeviceStatusController(syncDeviceStatusUseCase),

		AuthUseCase:          tuyaAuthUseCase,
		GetAllDevicesUseCase: tuyaGetAllDevicesUseCase,
		GetDeviceByIDUseCase: tuyaGetDeviceByIDUseCase,
		DeviceControlUseCase: tuyaDeviceControlBridge,
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
	routes.SetupTuyaControlRoutes(protected, m.CommandSwitchController, m.SendIRCommandController)

	// Sync Route
	protected.GET("/api/tuya/devices/sync", m.SyncDeviceStatusController.SyncStatus)
}
