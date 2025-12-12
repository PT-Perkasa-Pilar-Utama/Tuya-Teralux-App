package routes

import (
	"teralux_app/controllers"

	"github.com/gin-gonic/gin"
)

// SetupTuyaDeviceRoutes registers Tuya device routes
func SetupTuyaDeviceRoutes(
	router *gin.Engine,
	getAllDevicesController *controllers.TuyaGetAllDevicesController,
	getDeviceByIDController *controllers.TuyaGetDeviceByIDController,
) {
	api := router.Group("/api/tuya")
	{
		api.GET("/devices", getAllDevicesController.GetAllDevices)
		api.GET("/devices/:id", getDeviceByIDController.GetDeviceByID)
	}
}
