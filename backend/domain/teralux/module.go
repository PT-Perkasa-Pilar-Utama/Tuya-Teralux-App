package teralux

import (
	"teralux_app/domain/common/infrastructure"
	device "teralux_app/domain/teralux/controllers/device"
	device_status "teralux_app/domain/teralux/controllers/device_status"
	teralux "teralux_app/domain/teralux/controllers/teralux"
	"teralux_app/domain/teralux/repositories"
	"teralux_app/domain/teralux/routes"

	"teralux_app/domain/common/middlewares"
	device_usecases "teralux_app/domain/teralux/usecases/device"
	device_status_usecases "teralux_app/domain/teralux/usecases/device_status"
	teralux_usecases "teralux_app/domain/teralux/usecases/teralux"

	tuya_usecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TeraluxModule encapsulates Teralux domain components
type TeraluxModule struct {
	// Teralux Controllers
	CreateController   *teralux.CreateTeraluxController
	GetAllController   *teralux.GetAllTeraluxController
	GetByIDController  *teralux.GetTeraluxByIDController
	GetByMACController *teralux.GetTeraluxByMACController
	UpdateController   *teralux.UpdateTeraluxController
	DeleteController   *teralux.DeleteTeraluxController

	// Device Controllers
	CreateDeviceController          *device.CreateDeviceController
	GetAllDevicesController         *device.GetAllDevicesController
	GetDeviceByIDController         *device.GetDeviceByIDController
	GetDevicesByTeraluxIDController *device.GetDevicesByTeraluxIDController
	UpdateDeviceController          *device.UpdateDeviceController
	DeleteDeviceController          *device.DeleteDeviceController

	// DeviceStatus Controllers
	GetAllDeviceStatusesController        *device_status.GetAllDeviceStatusesController
	GetDeviceStatusByCodeController       *device_status.GetDeviceStatusByCodeController
	GetDeviceStatusesByDeviceIDController *device_status.GetDeviceStatusesByDeviceIDController
	UpdateDeviceStatusController          *device_status.UpdateDeviceStatusController
}

// NewTeraluxModule initializes the Teralux module
func NewTeraluxModule(
	badger *infrastructure.BadgerService,
	deviceRepository *repositories.DeviceRepository,
	tuyaAuthUC *tuya_usecases.TuyaAuthUseCase,
	tuyaGetDeviceUC *tuya_usecases.TuyaGetDeviceByIDUseCase,
) *TeraluxModule {
	// Repositories
	teraluxRepository := repositories.NewTeraluxRepository(badger)
	deviceStatusRepository := repositories.NewDeviceStatusRepository(badger)

	// Teralux Use Cases
	createTeraluxUseCase := teralux_usecases.NewCreateTeraluxUseCase(teraluxRepository)
	getAllTeraluxUseCase := teralux_usecases.NewGetAllTeraluxUseCase(teraluxRepository)
	getTeraluxByIDUseCase := teralux_usecases.NewGetTeraluxByIDUseCase(teraluxRepository, deviceRepository)
	getTeraluxByMACUseCase := teralux_usecases.NewGetTeraluxByMACUseCase(teraluxRepository)
	updateTeraluxUseCase := teralux_usecases.NewUpdateTeraluxUseCase(teraluxRepository)
	deleteTeraluxUseCase := teralux_usecases.NewDeleteTeraluxUseCase(teraluxRepository)

	// Device Use Cases
	createDeviceUseCase := device_usecases.NewCreateDeviceUseCase(deviceRepository, deviceStatusRepository, tuyaAuthUC, tuyaGetDeviceUC, teraluxRepository)
	getAllDevicesUseCase := device_usecases.NewGetAllDevicesUseCase(deviceRepository)
	getDeviceByIDUseCase := device_usecases.NewGetDeviceByIDUseCase(deviceRepository)
	getDevicesByTeraluxIDUseCase := device_usecases.NewGetDevicesByTeraluxIDUseCase(deviceRepository, teraluxRepository)
	updateDeviceUseCase := device_usecases.NewUpdateDeviceUseCase(deviceRepository)
	deleteDeviceUseCase := device_usecases.NewDeleteDeviceUseCase(deviceRepository, deviceStatusRepository, teraluxRepository)

	// Device Status Use Cases
	getDeviceStatusesByDeviceIDUseCase := device_status_usecases.NewGetDeviceStatusesByDeviceIDUseCase(deviceStatusRepository, deviceRepository)
	getAllDeviceStatusesUseCase := device_status_usecases.NewGetAllDeviceStatusesUseCase(deviceStatusRepository)
	getDeviceStatusByCodeUseCase := device_status_usecases.NewGetDeviceStatusByCodeUseCase(deviceStatusRepository, deviceRepository)
	updateDeviceStatusUseCase := device_status_usecases.NewUpdateDeviceStatusUseCase(deviceStatusRepository, deviceRepository)

	// Controllers
	return &TeraluxModule{
		CreateController:   teralux.NewCreateTeraluxController(createTeraluxUseCase),
		GetAllController:   teralux.NewGetAllTeraluxController(getAllTeraluxUseCase),
		GetByIDController:  teralux.NewGetTeraluxByIDController(getTeraluxByIDUseCase),
		GetByMACController: teralux.NewGetTeraluxByMACController(getTeraluxByMACUseCase),
		UpdateController:   teralux.NewUpdateTeraluxController(updateTeraluxUseCase),
		DeleteController:   teralux.NewDeleteTeraluxController(deleteTeraluxUseCase),

		CreateDeviceController:          device.NewCreateDeviceController(createDeviceUseCase),
		GetAllDevicesController:         device.NewGetAllDevicesController(getAllDevicesUseCase),
		GetDeviceByIDController:         device.NewGetDeviceByIDController(getDeviceByIDUseCase),
		GetDevicesByTeraluxIDController: device.NewGetDevicesByTeraluxIDController(getDevicesByTeraluxIDUseCase),
		UpdateDeviceController:          device.NewUpdateDeviceController(updateDeviceUseCase),
		DeleteDeviceController:          device.NewDeleteDeviceController(deleteDeviceUseCase),

		GetAllDeviceStatusesController:        device_status.NewGetAllDeviceStatusesController(getAllDeviceStatusesUseCase),
		GetDeviceStatusByCodeController:       device_status.NewGetDeviceStatusByCodeController(getDeviceStatusByCodeUseCase),
		GetDeviceStatusesByDeviceIDController: device_status.NewGetDeviceStatusesByDeviceIDController(getDeviceStatusesByDeviceIDUseCase),
		UpdateDeviceStatusController:          device_status.NewUpdateDeviceStatusController(updateDeviceStatusUseCase),
	}
}

// RegisterRoutes registers Teralux routes
func (m *TeraluxModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	// Public Group with API Key
	publicGroup := router.Group("/")
	publicGroup.Use(middlewares.ApiKeyMiddleware())

	routes.SetupTeraluxRoutes(
		publicGroup,
		protected,
		m.CreateController,
		m.GetAllController,
		m.GetByIDController,
		m.GetByMACController,
		m.UpdateController,
		m.DeleteController,

		m.CreateDeviceController,
		m.GetAllDevicesController,
		m.GetDeviceByIDController,
		m.GetDevicesByTeraluxIDController,
		m.UpdateDeviceController,
		m.DeleteDeviceController,

		m.GetAllDeviceStatusesController,
		m.GetDeviceStatusByCodeController,
		m.GetDeviceStatusesByDeviceIDController,
		m.UpdateDeviceStatusController,
	)
}
