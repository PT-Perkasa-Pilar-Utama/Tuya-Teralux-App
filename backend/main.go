package main

import (


	"github.com/gin-gonic/gin"

	"teralux_app/domain/common"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/common/middlewares"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux"
	"teralux_app/domain/tuya"


)

// @title           Teralux API
// @version         1.0
// @description     This is the API server for Teralux App.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name 01. Auth
// @tag.description Authentication endpoints

// @tag.name 02. Tuya
// @tag.description Tuya related endpoints

// @tag.name 03. Teralux
// @tag.description Teralux device management endpoints
func main() {
	utils.LoadConfig()

	// Initialize database connection
	_, err := infrastructure.InitDB()
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize database: %v", err)
	} else {
		defer infrastructure.CloseDB()
		utils.LogInfo("Database initialized successfully")
	}

	router := gin.Default()

	// Initialize BadgerDB
	badgerService, err := persistence.NewBadgerService("./tmp/badger")
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize BadgerDB: %v", err)
	} else {
		defer badgerService.Close()
	}

	// Initialize Modules
	commonModule := common.NewCommonModule(badgerService)
	tuyaModule := tuya.NewTuyaModule(badgerService)
	teraluxModule := teralux.NewTeraluxModule(badgerService)

	// Register Routes
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(middlewares.TuyaErrorMiddleware())

	// 1. Common Routes (Health, Cache)
	commonModule.RegisterRoutes(router, protected)

	// 2. Tuya Routes (Auth, Device Control)
	tuyaModule.RegisterRoutes(router, protected)

	// 3. Teralux Routes (CRUD)
	teraluxModule.RegisterRoutes(router, protected)
	
	utils.LogInfo("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		utils.LogInfo("Failed to start server: %v", err)
	}
}