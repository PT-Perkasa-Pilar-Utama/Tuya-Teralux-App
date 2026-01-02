package routes

import (
	device "teralux_app/domain/teralux/controllers/device"
	device_status "teralux_app/domain/teralux/controllers/device_status"
	teralux "teralux_app/domain/teralux/controllers/teralux"

	"github.com/gin-gonic/gin"
)


// SetupTeraluxRoutes registers endpoints for teralux CRUD operations
func SetupTeraluxRoutes(
	router gin.IRouter,
	createController *teralux.CreateTeraluxController,
	getAllController *teralux.GetAllTeraluxController,
	getByIDController *teralux.GetTeraluxByIDController,
	updateController *teralux.UpdateTeraluxController,
	deleteController *teralux.DeleteTeraluxController,

	createDeviceController *device.CreateDeviceController,
	getAllDevicesController *device.GetAllDevicesController,
	getDeviceByIDController *device.GetDeviceByIDController,
	updateDeviceController *device.UpdateDeviceController,
	deleteDeviceController *device.DeleteDeviceController,

	createDeviceStatusController *device_status.CreateDeviceStatusController,
	getAllDeviceStatusesController *device_status.GetAllDeviceStatusesController,
	getDeviceStatusByIDController *device_status.GetDeviceStatusByIDController,
	updateDeviceStatusController *device_status.UpdateDeviceStatusController,
	deleteDeviceStatusController *device_status.DeleteDeviceStatusController,
) {
	// Teralux Routes
	teraluxAPI := router.Group("/api/teralux")
	{
		// POST /api/teralux - Create a new teralux
		teraluxAPI.POST("", createController.CreateTeralux)

		// GET /api/teralux - Get all teralux
		teraluxAPI.GET("", getAllController.GetAllTeralux)

		// GET /api/teralux/:id - Get teralux by ID
		teraluxAPI.GET("/:id", getByIDController.GetTeraluxByID)

		// PUT /api/teralux/:id - Update teralux
		teraluxAPI.PUT("/:id", updateController.UpdateTeralux)

		// DELETE /api/teralux/:id - Delete teralux (soft delete)
		teraluxAPI.DELETE("/:id", deleteController.DeleteTeralux)
	}

	// Device Routes
	deviceAPI := router.Group("/api/devices")
	{
		deviceAPI.POST("", createDeviceController.CreateDevice)
		deviceAPI.GET("", getAllDevicesController.GetAllDevices)
		deviceAPI.GET("/:id", getDeviceByIDController.GetDeviceByID)
		deviceAPI.PUT("/:id", updateDeviceController.UpdateDevice)
		deviceAPI.DELETE("/:id", deleteDeviceController.DeleteDevice)
	}

	// Device Status Routes
	deviceStatusAPI := router.Group("/api/device-statuses")
	{
		deviceStatusAPI.POST("", createDeviceStatusController.CreateDeviceStatus)
		deviceStatusAPI.GET("", getAllDeviceStatusesController.GetAllDeviceStatuses)
		deviceStatusAPI.GET("/:id", getDeviceStatusByIDController.GetDeviceStatusByID)
		deviceStatusAPI.PUT("/:id", updateDeviceStatusController.UpdateDeviceStatus)
		deviceStatusAPI.DELETE("/:id", deleteDeviceStatusController.DeleteDeviceStatus)
	}
}

