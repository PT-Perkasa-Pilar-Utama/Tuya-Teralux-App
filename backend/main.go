package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/common"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/middlewares"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux"
	teralux_entities "teralux_app/domain/teralux/entities"
	teralux_repositories "teralux_app/domain/teralux/repositories"
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

// @tag.name 04. Devices
// @tag.description Teralux specific devices endpoints

// @tag.name 05. Device Statuses
// @tag.description Device status management endpoints

// @tag.name 06. Flush
// @tag.description Cache management endpoints

// @tag.name 07. Health
// @tag.description Health check endpoint
func main() {
	// CLI: Healthcheck
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://localhost:8080/health")
		if err != nil || resp.StatusCode != 200 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	utils.LoadConfig()

	// Handle Migrations (replacement for entrypoint.sh in distroless)
	RunMigrations()

	// Initialize database connection
	_, err := infrastructure.InitDB()
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize database: %v", err)
	} else {
		defer infrastructure.CloseDB()
		utils.LogInfo("Database initialized successfully")

		// Auto Migrate Entities
		if err := infrastructure.DB.AutoMigrate(
			&teralux_entities.Teralux{},
			&teralux_entities.Device{},
			&teralux_entities.DeviceStatus{},
		); err != nil {
			utils.LogInfo("Warning: Failed to auto-migrate entities: %v", err)
		} else {
			utils.LogInfo("Entities auto-migrated successfully")
		}
	}

	router := gin.Default()

	// Initialize Models & Repositories
	// Initialize BadgerDB
	badgerService, err := infrastructure.NewBadgerService("./tmp/badger")
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize BadgerDB: %v", err)
	} else {
		defer badgerService.Close()
	}

	// Shared Repositories
	deviceRepo := teralux_repositories.NewDeviceRepository(badgerService)

	// Initialize Modules
	commonModule := common.NewCommonModule(badgerService)
	tuyaModule := tuya.NewTuyaModule(badgerService)
	teraluxModule := teralux.NewTeraluxModule(badgerService, deviceRepo, tuyaModule.AuthUseCase, tuyaModule.GetDeviceByIDUseCase)
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

func RunMigrations() {
	if os.Getenv("DB_TYPE") != "mysql" || os.Getenv("AUTO_MIGRATE") != "true" {
		return
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Wait for DB
	utils.LogInfo("⏳ Waiting for MySQL to be ready...")
	for i := 0; i < 30; i++ {
		cmd := exec.Command("mysqladmin", "ping", "-h", dbHost, "--silent")
		if err := cmd.Run(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	utils.LogInfo("🔄 Running MySQL migrations...")
	cmd := exec.Command("/usr/local/bin/migrate", "-path", "/app/migrations", "-database", dsn, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		utils.LogInfo("❌ MySQL migrations failed: %v", err)
		os.Exit(1)
	}
	utils.LogInfo("✅ MySQL migrations completed")
}
