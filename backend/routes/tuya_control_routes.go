package routes

import (
	"teralux_app/controllers"

	"github.com/gin-gonic/gin"
)

// SetupTuyaControlRoutes registers Tuya device control routes
func SetupTuyaControlRoutes(router gin.IRouter, controller *controllers.TuyaDeviceControlController) {
	api := router.Group("/api/tuya")
	{
		// Send commands to device
		// URL: /api/tuya/devices/:id/commands
		// Method: POST
		// Headers:
		//    - Authorization: Bearer <token>
		// Param: id (string)
		// Body: {
		//    "commands": [
		//      { "code": "switch_1", "value": true }
		//    ]
		// }
		// Response: {
		//    "status": true,
		//    "message": "Command sent successfully",
		//    "data": { "success": true }
		// }
		api.POST("/devices/:id/commands", controller.SendCommand)

		// Send IR AC command
		// URL: /api/tuya/ir-ac/command
		// Method: POST
		// Headers:
		//    - Authorization: Bearer <token>
		// Body: {
		//    "infrared_id": "...",
		//    "remote_id": "...",
		//    "code": "...",
		//    "value": 1
		// }
		// Response: {
		//    "status": true,
		//    "message": "IR AC Command sent successfully",
		//    "data": { "success": true }
		// }
		api.POST("/ir-ac/command", controller.SendIRACCommand)
	}
}
