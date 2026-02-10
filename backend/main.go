package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/common"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/middlewares"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag"
	"teralux_app/domain/speech"
	"teralux_app/domain/teralux"
	teralux_entities "teralux_app/domain/teralux/entities"
	teralux_repositories "teralux_app/domain/teralux/repositories"
	"teralux_app/domain/tuya"
)

// @title           Teralux API
// @version         1.0
// @description     This is the API server for Teralux App. <br> For full documentation, visit <a href="/docs">/docs</a>.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath        /
// @schemes         http https
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

// @tag.name 08. Speech
// @tag.description Speech endpoints

// @tag.name 09. RAG
// @tag.description RAG endpoints

// @tag.name 10. Health
// @tag.description Health check endpoint
func main() {
	// CLI: Healthcheck
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		// Ensure config is loaded and use centralized PORT value
		utils.LoadConfig()
		port := utils.GetConfig().Port
		if port == "" {
			port = "8080"
		}
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://localhost:" + port + "/health")
		if err != nil || resp.StatusCode != 200 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	utils.LoadConfig()

	// Initialize database connection
	_, err := infrastructure.InitDB()
	if err != nil {
		utils.LogInfo("FATAL: Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer infrastructure.CloseDB()
	utils.LogInfo("Database initialized successfully")

	// Auto Migrate Entities
	if err := infrastructure.DB.AutoMigrate(
		&teralux_entities.Teralux{},
		&teralux_entities.Device{},
		&teralux_entities.DeviceStatus{},
	); err != nil {
		utils.LogInfo("FATAL: Failed to auto-migrate entities: %v", err)
		os.Exit(1)
	}
	utils.LogInfo("Entities auto-migrated successfully")

	router := gin.Default()
	router.Use(middlewares.CorsMiddleware())

	// Initialize Models & Repositories
	// Initialize BadgerDB
	badgerService, err := infrastructure.NewBadgerService("./tmp/badger")
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize BadgerDB: %v", err)
	} else {
		defer badgerService.Close()
	}

	// Initialize Vector DB
	vectorService := infrastructure.NewVectorService("./tmp/vector/store.json")

	// Shared Repositories
	deviceRepo := teralux_repositories.NewDeviceRepository(badgerService)

	// Initialize Modules
	commonModule := common.NewCommonModule(badgerService, vectorService)
	tuyaModule := tuya.NewTuyaModule(badgerService, vectorService)

	teraluxModule := teralux.NewTeraluxModule(badgerService, deviceRepo, tuyaModule.AuthUseCase, tuyaModule.GetDeviceByIDUseCase, tuyaModule.DeviceControlUseCase)
	// Register Routes
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware(tuyaModule.AuthUseCase))
	protected.Use(middlewares.TuyaErrorMiddleware())

	// 1. Common Routes (Health, Cache)
	commonModule.RegisterRoutes(router, protected)

	// 2. Tuya Routes (Auth, Device Control)
	tuyaModule.RegisterRoutes(router, protected)

	// 3. Teralux Routes (CRUD)
	teraluxModule.RegisterRoutes(router, protected)

	// 4. Speech & RAG Modules (migrated from stt-service)
	scfg := utils.GetConfig()
	// Log current log level for diagnostic purposes
	fmt.Printf("Application log level: %s\n", utils.GetCurrentLogLevelName())
	missing := []string{}
	if scfg.LLMProvider == "" {
		missing = append(missing, "LLM_PROVIDER")
	}
	// LLM_BASE_URL is required for Antigravity and Ollama, but not for direct Gemini (uses Google SDK/API default)
	if scfg.LLMProvider != "gemini" && scfg.LLMBaseURL == "" {
		missing = append(missing, "LLM_BASE_URL")
	}
	if scfg.LLMModel == "" {
		missing = append(missing, "LLM_MODEL")
	}
	if scfg.WhisperModelPath == "" {
		missing = append(missing, "WHISPER_MODEL_PATH")
	}
	if scfg.MaxFileSize == 0 {
		missing = append(missing, "MAX_FILE_SIZE_MB")
	}
	if scfg.Port == "" {
		missing = append(missing, "PORT")
	}
	if len(missing) > 0 {
		utils.LogError("FATAL: Speech/RAG config incomplete: %v", missing)
		os.Exit(1)
	} else {
		// Initialize RAG first as it's a dependency for Speech
		utils.LogInfo("Configuring LLM: Provider=%s, Model=%s", scfg.LLMProvider, scfg.LLMModel)
		ragUsecase := rag.InitModule(protected, scfg, badgerService, vectorService, tuyaModule.AuthUseCase)

		// Initialize Speech with RAG, Badger and Tuya Auth dependencies
		speech.InitModule(protected, scfg, badgerService, ragUsecase, tuyaModule.AuthUseCase)
	}

	// Register Health at the end so it appears last in Swagger
	router.GET("/health", commonModule.HealthController.CheckHealth)

	port := scfg.Port
	if port == "" {
		port = "8080"
	}
	utils.LogInfo("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		utils.LogInfo("Failed to start server: %v", err)
	}
}
