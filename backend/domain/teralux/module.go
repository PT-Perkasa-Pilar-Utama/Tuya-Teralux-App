package teralux

import (
	"teralux_app/domain/common/infrastructure/persistence"
	device "teralux_app/domain/teralux/controllers/device"
	device_status "teralux_app/domain/teralux/controllers/device_status"
	teralux "teralux_app/domain/teralux/controllers/teralux"
	"teralux_app/domain/teralux/repositories"
	"teralux_app/domain/teralux/routes"

	device_usecases "teralux_app/domain/teralux/usecases/device"
	device_status_usecases "teralux_app/domain/teralux/usecases/device_status"
	teralux_usecases "teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// TeraluxModule encapsulates Teralux domain components
type TeraluxModule struct {
	// Teralux Controllers
	CreateController    *teralux.CreateTeraluxController
	GetAllController    *teralux.GetAllTeraluxController
	GetByIDController   *teralux.GetTeraluxByIDController
	UpdateController    *teralux.UpdateTeraluxController
	DeleteController    *teralux.DeleteTeraluxController

	// Device Controllers
	CreateDeviceController  *device.CreateDeviceController
	GetAllDevicesController *device.GetAllDevicesController
	GetDeviceByIDController *device.GetDeviceByIDController
	UpdateDeviceController  *device.UpdateDeviceController
	DeleteDeviceController  *device.DeleteDeviceController

	// DeviceStatus Controllers
	CreateDeviceStatusController  *device_status.CreateDeviceStatusController
	GetAllDeviceStatusesController *device_status.GetAllDeviceStatusesController
	GetDeviceStatusByIDController *device_status.GetDeviceStatusByIDController
	UpdateDeviceStatusController  *device_status.UpdateDeviceStatusController
	DeleteDeviceStatusController  *device_status.DeleteDeviceStatusController
}

// NewTeraluxModule initializes the Teralux module
func NewTeraluxModule(badger *persistence.BadgerService) *TeraluxModule {
	// Repositories
	teraluxRepository := repositories.NewTeraluxRepository(badger)
	deviceRepository := repositories.NewDeviceRepository(badger)
	deviceStatusRepository := repositories.NewDeviceStatusRepository(badger)

	// Teralux Use Cases
	createTeraluxUseCase := teralux_usecases.NewCreateTeraluxUseCase(teraluxRepository)
	getAllTeraluxUseCase := teralux_usecases.NewGetAllTeraluxUseCase(teraluxRepository)
	getTeraluxByIDUseCase := teralux_usecases.NewGetTeraluxByIDUseCase(teraluxRepository)
	updateTeraluxUseCase := teralux_usecases.NewUpdateTeraluxUseCase(teraluxRepository)
	deleteTeraluxUseCase := teralux_usecases.NewDeleteTeraluxUseCase(teraluxRepository)

	// Device Use Cases
	createDeviceUseCase := device_usecases.NewCreateDeviceUseCase(deviceRepository)
	getAllDevicesUseCase := device_usecases.NewGetAllDevicesUseCase(deviceRepository)
	getDeviceByIDUseCase := device_usecases.NewGetDeviceByIDUseCase(deviceRepository)
	updateDeviceUseCase := device_usecases.NewUpdateDeviceUseCase(deviceRepository)
	deleteDeviceUseCase := device_usecases.NewDeleteDeviceUseCase(deviceRepository)

	// Device Status Use Cases
	createDeviceStatusUseCase := device_status_usecases.NewCreateDeviceStatusUseCase(deviceStatusRepository)
	getAllDeviceStatusesUseCase := device_status_usecases.NewGetAllDeviceStatusesUseCase(deviceStatusRepository)
	getDeviceStatusByIDUseCase := device_status_usecases.NewGetDeviceStatusByIDUseCase(deviceStatusRepository)
	updateDeviceStatusUseCase := device_status_usecases.NewUpdateDeviceStatusUseCase(deviceStatusRepository)
	deleteDeviceStatusUseCase := device_status_usecases.NewDeleteDeviceStatusUseCase(deviceStatusRepository)

	// Controllers
	return &TeraluxModule{
		CreateController:    teralux.NewCreateTeraluxController(createTeraluxUseCase),
		GetAllController:    teralux.NewGetAllTeraluxController(getAllTeraluxUseCase),
		GetByIDController:   teralux.NewGetTeraluxByIDController(getTeraluxByIDUseCase),
		UpdateController:    teralux.NewUpdateTeraluxController(updateTeraluxUseCase),
		DeleteController:    teralux.NewDeleteTeraluxController(deleteTeraluxUseCase),
        
		CreateDeviceController:  device.NewCreateDeviceController(createDeviceUseCase),
		GetAllDevicesController: device.NewGetAllDevicesController(getAllDevicesUseCase),
		GetDeviceByIDController: device.NewGetDeviceByIDController(getDeviceByIDUseCase),
		UpdateDeviceController:  device.NewUpdateDeviceController(updateDeviceUseCase),
		DeleteDeviceController:  device.NewDeleteDeviceController(deleteDeviceUseCase),

		CreateDeviceStatusController:  device_status.NewCreateDeviceStatusController(createDeviceStatusUseCase),
		GetAllDeviceStatusesController: device_status.NewGetAllDeviceStatusesController(getAllDeviceStatusesUseCase),
		GetDeviceStatusByIDController: device_status.NewGetDeviceStatusByIDController(getDeviceStatusByIDUseCase),
		UpdateDeviceStatusController:  device_status.NewUpdateDeviceStatusController(updateDeviceStatusUseCase),
		DeleteDeviceStatusController:  device_status.NewDeleteDeviceStatusController(deleteDeviceStatusUseCase),
	}
}

// RegisterRoutes registers Teralux routes
func (m *TeraluxModule) RegisterRoutes(protected *gin.RouterGroup) {
	routes.SetupTeraluxRoutes(
		protected,
		m.CreateController,
		m.GetAllController,
		m.GetByIDController,
		m.UpdateController,
		m.DeleteController,
		
		m.CreateDeviceController,
		m.GetAllDevicesController,
		m.GetDeviceByIDController,
		m.UpdateDeviceController,
		m.DeleteDeviceController,
		
		m.CreateDeviceStatusController,
		m.GetAllDeviceStatusesController,
		m.GetDeviceStatusByIDController,
		m.UpdateDeviceStatusController,
		m.DeleteDeviceStatusController,
	)
}
