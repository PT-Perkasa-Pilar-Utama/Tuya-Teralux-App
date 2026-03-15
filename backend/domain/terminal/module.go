package terminal

import (
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	device "sensio/domain/terminal/device/controllers"
	device_repositories "sensio/domain/terminal/device/repositories"
	device_status "sensio/domain/terminal/device_status/controllers"
	device_status_repositories "sensio/domain/terminal/device_status/repositories"
	terminal "sensio/domain/terminal/terminal/controllers"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	"sensio/domain/terminal/terminal/routes"

	"sensio/domain/common/middlewares"
	device_usecases "sensio/domain/terminal/device/usecases"
	device_status_usecases "sensio/domain/terminal/device_status/usecases"
	terminal_services "sensio/domain/terminal/terminal/services"
	terminal_usecases "sensio/domain/terminal/terminal/usecases"

	tuya_usecases "sensio/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TerminalModule encapsulates Terminal domain components
type TerminalModule struct {
	// Terminal Controllers
	CreateController             *terminal.CreateTerminalController
	GetAllController             *terminal.GetAllTerminalController
	GetByIDController            *terminal.GetTerminalByIDController
	GetByMACController           *terminal.GetTerminalByMACController
	GetMQTTCredentialsController *terminal.GetMQTTCredentialsController
	UpdateController             *terminal.UpdateTerminalController
	DeleteController             *terminal.DeleteTerminalController

	// Device Controllers
	CreateDeviceController           *device.CreateDeviceController
	GetAllDevicesController          *device.GetAllDevicesController
	GetDeviceByIDController          *device.GetDeviceByIDController
	GetDevicesByTerminalIDController *device.GetDevicesByTerminalIDController
	UpdateDeviceController           *device.UpdateDeviceController
	DeleteDeviceController           *device.DeleteDeviceController

	// DeviceStatus Controllers
	GetAllDeviceStatusesController        *device_status.GetAllDeviceStatusesController
	GetDeviceStatusByCodeController       *device_status.GetDeviceStatusByCodeController
	GetDeviceStatusesByDeviceIDController *device_status.GetDeviceStatusesByDeviceIDController
	UpdateDeviceStatusController          *device_status.UpdateDeviceStatusController
}

// NewTerminalModule initializes the Terminal module
func NewTerminalModule(
	badger *infrastructure.BadgerService,
	deviceRepository *device_repositories.DeviceRepository,
	tuyaAuthUC tuya_usecases.TuyaAuthUseCase,
	tuyaGetDeviceUC *tuya_usecases.TuyaGetDeviceByIDUseCase,
	tuyaDeviceControlUC device_status_usecases.TuyaDeviceControlExecutor,
) *TerminalModule {
	// Services
	terminalExternalService := terminal_services.NewTerminalExternalService()

	// MQTT Auth Service client (points to EMQX Auth Service / Rust)
	cfg := utils.GetConfig()
	mqttAuthClient := terminal_services.NewMqttAuthClient(cfg.EmqxAuthBaseURL, cfg.EmqxAuthApiKey)

	// Repositories
	terminalRepository := terminal_repositories.NewTerminalRepository(badger)
	deviceStatusRepository := device_status_repositories.NewDeviceStatusRepository(badger)

	// Terminal Use Cases
	createTerminalUseCase := terminal_usecases.NewCreateTerminalUseCase(terminalRepository, terminalExternalService, mqttAuthClient)
	getAllTerminalUseCase := terminal_usecases.NewGetAllTerminalUseCase(terminalRepository)
	getTerminalByIDUseCase := terminal_usecases.NewGetTerminalByIDUseCase(terminalRepository, deviceRepository)
	getTerminalByMACUseCase := terminal_usecases.NewGetTerminalByMACUseCase(terminalRepository, mqttAuthClient)
	updateTerminalUseCase := terminal_usecases.NewUpdateTerminalUseCase(terminalRepository)
	deleteTerminalUseCase := terminal_usecases.NewDeleteTerminalUseCase(terminalRepository, mqttAuthClient)

	// Device Use Cases
	createDeviceUseCase := device_usecases.NewCreateDeviceUseCase(deviceRepository, deviceStatusRepository, tuyaAuthUC, tuyaGetDeviceUC, terminalRepository)
	getAllDevicesUseCase := device_usecases.NewGetAllDevicesUseCase(deviceRepository)
	getDeviceByIDUseCase := device_usecases.NewGetDeviceByIDUseCase(deviceRepository)
	getDevicesByTerminalIDUseCase := device_usecases.NewGetDevicesByTerminalIDUseCase(deviceRepository, terminalRepository)
	updateDeviceUseCase := device_usecases.NewUpdateDeviceUseCase(deviceRepository, terminalRepository)
	deleteDeviceUseCase := device_usecases.NewDeleteDeviceUseCase(deviceRepository, deviceStatusRepository, terminalRepository)

	// Device Status Use Cases
	getDeviceStatusesByDeviceIDUseCase := device_status_usecases.NewGetDeviceStatusesByDeviceIDUseCase(deviceStatusRepository, deviceRepository)
	getAllDeviceStatusesUseCase := device_status_usecases.NewGetAllDeviceStatusesUseCase(deviceStatusRepository)
	getDeviceStatusByCodeUseCase := device_status_usecases.NewGetDeviceStatusByCodeUseCase(deviceStatusRepository, deviceRepository)
	updateDeviceStatusUseCase := device_status_usecases.NewUpdateDeviceStatusUseCase(deviceStatusRepository, deviceRepository, tuyaDeviceControlUC)

	// Controllers
	return &TerminalModule{
		CreateController:             terminal.NewCreateTerminalController(createTerminalUseCase),
		GetAllController:             terminal.NewGetAllTerminalController(getAllTerminalUseCase),
		GetByIDController:            terminal.NewGetTerminalByIDController(getTerminalByIDUseCase),
		GetByMACController:           terminal.NewGetTerminalByMACController(getTerminalByMACUseCase),
		GetMQTTCredentialsController: terminal.NewGetMQTTCredentialsController(mqttAuthClient),
		UpdateController:             terminal.NewUpdateTerminalController(updateTerminalUseCase),
		DeleteController:             terminal.NewDeleteTerminalController(deleteTerminalUseCase),

		CreateDeviceController:           device.NewCreateDeviceController(createDeviceUseCase),
		GetAllDevicesController:          device.NewGetAllDevicesController(getAllDevicesUseCase),
		GetDeviceByIDController:          device.NewGetDeviceByIDController(getDeviceByIDUseCase),
		GetDevicesByTerminalIDController: device.NewGetDevicesByTerminalIDController(getDevicesByTerminalIDUseCase),
		UpdateDeviceController:           device.NewUpdateDeviceController(updateDeviceUseCase),
		DeleteDeviceController:           device.NewDeleteDeviceController(deleteDeviceUseCase),

		GetAllDeviceStatusesController:        device_status.NewGetAllDeviceStatusesController(getAllDeviceStatusesUseCase),
		GetDeviceStatusByCodeController:       device_status.NewGetDeviceStatusByCodeController(getDeviceStatusByCodeUseCase),
		GetDeviceStatusesByDeviceIDController: device_status.NewGetDeviceStatusesByDeviceIDController(getDeviceStatusesByDeviceIDUseCase),
		UpdateDeviceStatusController:          device_status.NewUpdateDeviceStatusController(updateDeviceStatusUseCase),
	}
}

// RegisterRoutes registers Terminal routes
func (m *TerminalModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Public Group with API Key
	publicGroup := router.Group("/")
	publicGroup.Use(middlewares.ApiKeyMiddleware())

	routes.SetupTerminalRoutes(
		publicGroup,
		protected,
		m.CreateController,
		m.GetAllController,
		m.GetByIDController,
		m.GetByMACController,
		m.GetMQTTCredentialsController,
		m.UpdateController,
		m.DeleteController,

		m.CreateDeviceController,
		m.GetAllDevicesController,
		m.GetDeviceByIDController,
		m.GetDevicesByTerminalIDController,
		m.UpdateDeviceController,
		m.DeleteDeviceController,

		m.GetAllDeviceStatusesController,
		m.GetDeviceStatusByCodeController,
		m.GetDeviceStatusesByDeviceIDController,
		m.UpdateDeviceStatusController,
	)
}
