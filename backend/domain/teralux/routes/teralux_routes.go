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
	getDevicesByTeraluxIDController *device.GetDevicesByTeraluxIDController,
	updateDeviceController *device.UpdateDeviceController,
	deleteDeviceController *device.DeleteDeviceController,

	getAllDeviceStatusesController *device_status.GetAllDeviceStatusesController,
	getDeviceStatusByCodeController *device_status.GetDeviceStatusByCodeController,
	getDeviceStatusesByDeviceIDController *device_status.GetDeviceStatusesByDeviceIDController,
	updateDeviceStatusController *device_status.UpdateDeviceStatusController,
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
		deviceAPI.GET("/teralux/:teralux_id", getDevicesByTeraluxIDController.GetDevicesByTeraluxID)
		deviceAPI.GET("/:id", getDeviceByIDController.GetDeviceByID)
		deviceAPI.PUT("/:id", updateDeviceController.UpdateDevice)
		deviceAPI.DELETE("/:id", deleteDeviceController.DeleteDevice)
	}

	// Device Status Routes (Protected)
	// GET /api/devices/statuses - Get all statuses (Scenario 1)
	protectedRouter.GET("/api/devices/statuses", getAllDeviceStatusesController.GetAllDeviceStatuses)

	// GET /api/devices/:id/statuses - Get statuses by device ID (Scenario 2)
	protectedRouter.GET("/api/devices/:id/statuses", getDeviceStatusesByDeviceIDController.GetDeviceStatusesByDeviceID)

	// GET /api/devices/:id/statuses/:code - Get status by code (Scenario 4)
	protectedRouter.GET("/api/devices/:id/statuses/:code", getDeviceStatusByCodeController.GetDeviceStatusByCode)

	// PUT /api/devices/:id/status - Update device status (Scenario 1)
	protectedRouter.PUT("/api/devices/:id/status", updateDeviceStatusController.UpdateDeviceStatus)
}
