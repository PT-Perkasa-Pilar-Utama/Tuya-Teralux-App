package routes

import (
	"sensio/domain/common/controllers"

	"github.com/gin-gonic/gin"
)

// SetupDeviceInfoExternalRoutes registers endpoints for fetching Big API data.
//
// param rg The router group to attach the routes to.
// param controller The controller handling Big API operations.
func SetupDeviceInfoExternalRoutes(rg *gin.RouterGroup, controller *controllers.DeviceInfoExternalController) {
	bigGroup := rg.Group("/api/big")
	{
		// GET /api/big/device/:mac_address
		// Fetches booking details and device info by MAC address
		bigGroup.GET("/device/:mac_address", controller.GetDeviceInfo)
	}
}
