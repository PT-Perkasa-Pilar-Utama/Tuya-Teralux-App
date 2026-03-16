package tuya

import (
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/middlewares"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	device_repositories "sensio/domain/terminal/device/repositories"
	"sensio/domain/tuya/controllers"
	"sensio/domain/tuya/routes"
	"sensio/domain/tuya/services"
	"sensio/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaModule encapsulates Tuya domain components
type TuyaModule struct {
	AuthController          *controllers.TuyaAuthController
	GetAllDevicesController *controllers.TuyaGetAllDevicesController
	GetDeviceByIDController *controllers.TuyaGetDeviceByIDController
	CommandSwitchController *controllers.TuyaCommandSwitchController
	SendIRCommandController *controllers.TuyaSendIRCommandController
	SensorController        *controllers.TuyaSensorController

	// Exported Use Cases for other domains
	AuthUseCase          usecases.TuyaAuthUseCase
	GetAllDevicesUseCase usecases.TuyaGetAllDevicesUseCase
	GetDeviceByIDUseCase *usecases.TuyaGetDeviceByIDUseCase
	DeviceControlUseCase usecases.TuyaDeviceControlExecutor
}

// NewTuyaModule initializes the Tuya module
func NewTuyaModule(badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, deviceRepo *device_repositories.DeviceRepository, terminalRepo *terminal_repositories.TerminalRepository) *TuyaModule {
	// Services
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaDeviceService := services.NewTuyaDeviceService()

	// Use Cases
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)
	deviceStateUseCase := usecases.NewDeviceStateUseCase(badger)

	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService, deviceStateUseCase, badger, vectorSvc, deviceRepo, terminalRepo)
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService, deviceStateUseCase)
	tuyaCommandSwitchUseCase := usecases.NewTuyaCommandSwitchUseCase(tuyaDeviceService, deviceStateUseCase)
	tuyaSendIRCommandUseCase := usecases.NewTuyaSendIRCommandUseCase(tuyaDeviceService, deviceStateUseCase)

	// Bridge for shared executor
	tuyaDeviceControlBridge := usecases.NewTuyaDeviceControlBridge(tuyaCommandSwitchUseCase, tuyaSendIRCommandUseCase, badger)

	tuyaSensorUseCase := usecases.NewTuyaSensorUseCase(tuyaGetDeviceByIDUseCase)

	// Controllers
	return &TuyaModule{
		AuthController:          controllers.NewTuyaAuthController(tuyaAuthUseCase),
		GetAllDevicesController: controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase),
		GetDeviceByIDController: controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase),
		CommandSwitchController: controllers.NewTuyaCommandSwitchController(tuyaCommandSwitchUseCase),
		SendIRCommandController: controllers.NewTuyaSendIRCommandController(tuyaSendIRCommandUseCase),
		SensorController:        controllers.NewTuyaSensorController(tuyaSensorUseCase),

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
}
