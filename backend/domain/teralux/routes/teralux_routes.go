package routes

import (
	device "teralux_app/domain/teralux/controllers/device"
	device_status "teralux_app/domain/teralux/controllers/device_status"
	teralux "teralux_app/domain/teralux/controllers/teralux"

	"github.com/gin-gonic/gin"
)

// SetupTeraluxRoutes registers endpoints for teralux CRUD operations
func SetupTeraluxRoutes(
	publicRouter gin.IRouter,
	protectedRouter gin.IRouter,
	createController *teralux.CreateTeraluxController,
	getAllController *teralux.GetAllTeraluxController,
	getByIDController *teralux.GetTeraluxByIDController,
	getByMACController *teralux.GetTeraluxByMACController,
	updateController *teralux.UpdateTeraluxController,
	deleteController *teralux.DeleteTeraluxController,

	createDeviceController *device.CreateDeviceController,
	getAllDevicesController *device.GetAllDevicesController,
	getDeviceByIDController *device.GetDeviceByIDController,
	updateDeviceController *device.UpdateDeviceController,
	deleteDeviceController *device.DeleteDeviceController,

	createDeviceStatusController *device_status.CreateDeviceStatusController,
	getAllDeviceStatusesController *device_status.GetAllDeviceStatusesController,
	getDeviceStatusByCodeController *device_status.GetDeviceStatusByCodeController,
	updateDeviceStatusController *device_status.UpdateDeviceStatusController,
	deleteDeviceStatusController *device_status.DeleteDeviceStatusController,
) {
	// Public Teralux Routes (Registration and Check)
	teraluxPublicAPI := publicRouter.Group("/api/teralux")
	{
		// POST /api/teralux - Create a new teralux (Public with API Key)
		teraluxPublicAPI.POST("", createController.CreateTeralux)

		// GET /api/teralux - Get all teralux (Public with API Key for check)
		teraluxPublicAPI.GET("", getAllController.GetAllTeralux)

		// GET /api/teralux/mac/:mac - Get teralux by MAC address (Public with API Key)
		teraluxPublicAPI.GET("/mac/:mac", getByMACController.GetTeraluxByMAC)
	}

	// Protected Teralux Routes
	teraluxProtectedAPI := protectedRouter.Group("/api/teralux")
	{
		// GET /api/teralux/:id - Get teralux by ID
		teraluxProtectedAPI.GET("/:id", getByIDController.GetTeraluxByID)

		// PUT /api/teralux/:id - Update teralux
		teraluxProtectedAPI.PUT("/:id", updateController.UpdateTeralux)

		// DELETE /api/teralux/:id - Delete teralux (soft delete)
		teraluxProtectedAPI.DELETE("/:id", deleteController.DeleteTeralux)
	}

	// Device Routes (Protected)
	deviceAPI := protectedRouter.Group("/api/devices")
	{
		deviceAPI.POST("", createDeviceController.CreateDevice)
		deviceAPI.GET("", getAllDevicesController.GetAllDevices)
		deviceAPI.GET("/:id", getDeviceByIDController.GetDeviceByID)
		deviceAPI.PUT("/:id", updateDeviceController.UpdateDevice)
		deviceAPI.DELETE("/:id", deleteDeviceController.DeleteDevice)
	}

	// Device Status Routes (Protected)
	deviceStatusAPI := protectedRouter.Group("/api/device-statuses")
	{
		deviceStatusAPI.POST("", createDeviceStatusController.CreateDeviceStatus)
		deviceStatusAPI.GET("", getAllDeviceStatusesController.GetAllDeviceStatuses)
		deviceStatusAPI.GET("/:deviceId/:code", getDeviceStatusByCodeController.GetDeviceStatusByCode)
		deviceStatusAPI.PUT("/:deviceId/:code", updateDeviceStatusController.UpdateDeviceStatus)
		deviceStatusAPI.DELETE("/:deviceId/:code", deleteDeviceStatusController.DeleteDeviceStatus)
	}
}
