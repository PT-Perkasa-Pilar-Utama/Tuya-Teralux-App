package routes

import (
	device "sensio/domain/terminal/device/controllers"
	device_status "sensio/domain/terminal/device_status/controllers"
	terminal "sensio/domain/terminal/terminal/controllers"

	"github.com/gin-gonic/gin"
)

// SetupTerminalRoutes registers endpoints for terminal CRUD operations
func SetupTerminalRoutes(
	publicRouter gin.IRouter,
	protectedRouter gin.IRouter,
	createController *terminal.CreateTerminalController,
	getAllController *terminal.GetAllTerminalController,
	getByIDController *terminal.GetTerminalByIDController,
	getByMACController *terminal.GetTerminalByMACController,
	getMqttCredentialsController *terminal.GetMQTTCredentialsController,
	updateController *terminal.UpdateTerminalController,
	deleteController *terminal.DeleteTerminalController,

	createDeviceController *device.CreateDeviceController,
	getAllDevicesController *device.GetAllDevicesController,
	getDeviceByIDController *device.GetDeviceByIDController,
	getDevicesByTerminalIDController *device.GetDevicesByTerminalIDController,
	updateDeviceController *device.UpdateDeviceController,
	deleteDeviceController *device.DeleteDeviceController,

	getAllDeviceStatusesController *device_status.GetAllDeviceStatusesController,
	getDeviceStatusByCodeController *device_status.GetDeviceStatusByCodeController,
	getDeviceStatusesByDeviceIDController *device_status.GetDeviceStatusesByDeviceIDController,
	updateDeviceStatusController *device_status.UpdateDeviceStatusController,
) {
	// Public Terminal Routes (Bootstrap: Registration and Check)
	terminalPublicAPI := publicRouter.Group("/api/terminal")
	{
		// POST /api/terminal - Create a new terminal (API Key for bootstrap)
		terminalPublicAPI.POST("", createController.CreateTerminal)

		// GET /api/terminal/mac/:mac - Get terminal by MAC address (API Key for bootstrap)
		terminalPublicAPI.GET("/mac/:mac", getByMACController.GetTerminalByMAC)
	}

	// Protected Terminal Routes (Operational)
	terminalProtectedAPI := protectedRouter.Group("/api/terminal")
	{
		// GET /api/terminal - Get all terminals (Bearer)
		terminalProtectedAPI.GET("", getAllController.GetAllTerminal)

		// GET /api/terminal/:id - Get terminal by ID (Bearer)
		terminalProtectedAPI.GET("/:id", getByIDController.GetTerminalByID)

		// PUT /api/terminal/:id - Update terminal (Bearer)
		terminalProtectedAPI.PUT("/:id", updateController.UpdateTerminal)

		// DELETE /api/terminal/:id - Delete terminal (Bearer)
		terminalProtectedAPI.DELETE("/:id", deleteController.DeleteTerminal)
	}

	// Protected MQTT Routes
	mqttProtectedAPI := protectedRouter.Group("/api/mqtt")
	{
		// GET /api/mqtt/users/:username - Get MQTT credentials
		mqttProtectedAPI.GET("/users/:username", getMqttCredentialsController.GetMQTTCredentials)
	}

	// Device Routes (Protected)
	deviceAPI := protectedRouter.Group("/api/devices")
	{
		deviceAPI.POST("", createDeviceController.CreateDevice)
		deviceAPI.GET("", getAllDevicesController.GetAllDevices)
		deviceAPI.GET("/terminal/:terminal_id", getDevicesByTerminalIDController.GetDevicesByTerminalID)
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
