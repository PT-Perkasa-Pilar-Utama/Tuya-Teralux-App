package main

import (
	"log"
	"teralux_app/controllers"
	"teralux_app/services"
	"teralux_app/usecases"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	utils.LoadConfig()

	// Initialize Gin router
	router := gin.Default()

	// Initialize dependency chain: service -> usecase -> controller
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)

	tuyaDeviceService := services.NewTuyaDeviceService()

	// Initialize Get All Devices chain
	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService)
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService)
	tuyaDeviceControlUseCase := usecases.NewTuyaDeviceControlUseCase(tuyaDeviceService)

	tuyaAuthController := controllers.NewTuyaAuthController(tuyaAuthUseCase)
	tuyaGetAllDevicesController := controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase)
	tuyaGetDeviceByIDController := controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase)
	tuyaDeviceControlController := controllers.NewTuyaDeviceControlController(tuyaDeviceControlUseCase)

	// Routes
	api := router.Group("/api")
	{
		api.POST("/tuya/auth", tuyaAuthController.Authenticate)
		api.GET("/tuya/devices", tuyaGetAllDevicesController.GetAllDevices)
		api.GET("/tuya/devices/:id", tuyaGetDeviceByIDController.GetDeviceByID)
		api.POST("/tuya/devices/:id/commands", tuyaDeviceControlController.SendCommand)
		api.POST("/tuya/ir-ac/command", tuyaDeviceControlController.SendIRACCommand)
	}
	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
