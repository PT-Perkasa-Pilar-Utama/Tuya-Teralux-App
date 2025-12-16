package routes

import (
	"teralux_app/controllers"

	"github.com/gin-gonic/gin"
)																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																			

// SetupTuyaDeviceRoutes registers Tuya device routes
func SetupTuyaDeviceRoutes(
	router gin.IRouter,
	getAllDevicesController *controllers.TuyaGetAllDevicesController,
	getDeviceByIDController *controllers.TuyaGetDeviceByIDController,
) {
	// Group: /api/tuya
	api := router.Group("/api/tuya")
	{																																																																																																																																																						
		// Get all devices
		// URL: /api/tuya/devices
		// Method: GET
		// Headers:
		//    - Authorization: Bearer <token>
		// Response: {
		//    "status": true,
		//    "message": "Devices fetched successfully",
		//    "data": {
		//      "devices": [
		//        {
		//          "id": "...",
		//          "name": "...",
		//          "category": "...",
		//          "product_name": "...",
		//          "online": true,
		//          "icon": "...",
		//          "status": [
		//            { "code": "switch_1", "value": true }
		//          ]
		//        }
		//      ],
		//      "total": 1
		//    }
		// }
		api.GET("/devices", getAllDevicesController.GetAllDevices)

		// Get device by ID
		// URL: /api/tuya/devices/:id
		// Method: GET
		// Headers: 
		//    - Authorization: Bearer <token>
		// Param: id (string)
		// Response: {
		//    "status": true,
		//    "message": "Device fetched successfully",
		//    "data": {
		//      "device": {
		//          "id": "...",
		//          "name": "...",
		//          ...
		//      }
		//    }
		// }
		api.GET("/devices/:id", getDeviceByIDController.GetDeviceByID)

	}
}
