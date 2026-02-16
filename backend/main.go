package main

// Trigger documentation refresh build

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
	"teralux_app/domain/recordings"
	recordings_entities "teralux_app/domain/recordings/entities"
	"teralux_app/domain/scene"
	scene_entities "teralux_app/domain/scene/entities"
	"teralux_app/domain/speech"
	"teralux_app/domain/teralux"
	teralux_entities "teralux_app/domain/teralux/entities"
	teralux_repositories "teralux_app/domain/teralux/repositories"
	"teralux_app/domain/tuya"
)

// @title           Sensio API
// @version         1.0
// @description     This is the API server for Sensio App. <br> For full documentation, visit <a href="/docs">/docs</a>.
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

// @tag.name 03. Scenes
// @tag.description Scene management and control endpoints

// @tag.name 04. Speech
// @tag.description Speech endpoints

// @tag.name 05. RAG
// @tag.description RAG endpoints

// @tag.name 06. Recordings
// @tag.description Recordings management endpoints

// @tag.name 07. Flush
// @tag.description Cache management endpoints

// @tag.name 08. Health
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
	defer func() {
		_ = infrastructure.CloseDB()
	}()
	utils.LogInfo("Database initialized successfully")

	// Auto Migrate Entities
	if err := infrastructure.DB.AutoMigrate(
		&teralux_entities.Teralux{},
		&teralux_entities.Device{},
		&teralux_entities.DeviceStatus{},
		&scene_entities.Scene{},
		&recordings_entities.Recording{},
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

	// Initialize MQTT Service
	mqttService := infrastructure.NewMqttService(utils.GetConfig())
	if err := mqttService.Connect(); err != nil {
		utils.LogError("Warning: Failed to connect to MQTT: %v", err)
	} else {
		defer mqttService.Close()
	}

	// Shared Repositories
	deviceRepo := teralux_repositories.NewDeviceRepository(badgerService)
	teraluxRepo := teralux_repositories.NewTeraluxRepository(badgerService)

	// Initialize Modules
	commonModule := common.NewCommonModule(badgerService, vectorService, mqttService)
	tuyaModule := tuya.NewTuyaModule(badgerService, vectorService, deviceRepo, teraluxRepo)

	teraluxModule := teralux.NewTeraluxModule(badgerService, deviceRepo, tuyaModule.AuthUseCase, tuyaModule.GetDeviceByIDUseCase, tuyaModule.DeviceControlUseCase)
	// Register Routes
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware(tuyaModule.AuthUseCase))
	protected.Use(middlewares.TuyaErrorMiddleware())

	// Static File Serving (for audio uploads)
	// Access via: /uploads/audio/filename.ext
	router.Static("/uploads", "./uploads")

	// 1. Common Routes (Health, Cache)
	commonModule.RegisterRoutes(router, protected)

	// 2. Tuya Routes (Auth, Device Control)
	tuyaModule.RegisterRoutes(router, protected)

	// 3. Teralux Routes (CRUD)
	teraluxModule.RegisterRoutes(router, protected)

	// 4. Recordings Module
	recordingsModule := recordings.NewRecordingsModule(badgerService)
	recordingsModule.RegisterRoutes(router, protected)

	// 5. Speech & RAG Modules (migrated from stt-service)
	scfg := utils.GetConfig()
	// Log current log level for diagnostic purposes
	fmt.Printf("Application log level: %s\n", utils.GetCurrentLogLevelName())
	missing := []string{}
	if scfg.LLMProvider == "" {
		missing = append(missing, "LLM_PROVIDER")
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
		ragUsecase := rag.InitModule(protected, scfg, badgerService, vectorService, tuyaModule.AuthUseCase, tuyaModule.DeviceControlUseCase)

		// Initialize Speech with RAG, Badger and Tuya Auth dependencies
		speech.InitModule(protected, scfg, badgerService, ragUsecase, tuyaModule.AuthUseCase, mqttService, recordingsModule.SaveRecordingUseCase)

		// 6. Scene Module
		sceneModule := scene.NewSceneModule(infrastructure.DB, tuyaModule.DeviceControlUseCase, mqttService)
		sceneModule.RegisterRoutes(protected)
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
