package main

// Trigger documentation refresh build

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"sensio/domain/common"
	common_entities "sensio/domain/common/entities"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/middlewares"
	"sensio/domain/common/services"
	"sensio/domain/common/utils"
	"sensio/domain/mail"
	"sensio/domain/models"
	models_v1 "sensio/domain/models-v1"
	"sensio/domain/recordings"
	recordings_entities "sensio/domain/recordings/entities"
	"sensio/domain/scene"
	scene_entities "sensio/domain/scene/entities"
	"sensio/domain/terminal"
	device_entities "sensio/domain/terminal/device/entities"
	device_repositories "sensio/domain/terminal/device/repositories"
	terminal_entities "sensio/domain/terminal/terminal/entities"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	"sensio/domain/tuya"
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
// @description Enter JWT token only (without "Bearer " prefix)

// @tag.name 01. Tuya
// @tag.description Tuya authentication and device control endpoints

// @tag.name 02. Terminal
// @tag.description Terminal and device management endpoints

// @tag.name 03. Scenes
// @tag.description Scene management and control endpoints

// @tag.name 04. Models
// @tag.description AI Model access endpoints (Speech, RAG, Pipeline) - Unified domain

// @tag.name 05. Models-v1
// @tag.description New AI Model endpoints (Whisper, RAG, Pipeline) - v1 API

// @tag.name 06. Recordings
// @tag.description Recordings management endpoints

// @tag.name 07. Mail
// @tag.description Mail service endpoints

// @tag.name 08. Common
// @tag.description Common endpoints (Health, Cache, External APIs)
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
		resp, err := client.Get("http://localhost:" + port + "/api/health")
		if err != nil || resp.StatusCode != 200 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := run(); err != nil {
		utils.LogError("FATAL: %v", err)
		os.Exit(1)
	}
}

func run() error {
	utils.LoadConfig()

	// Initialize database connection
	_, err := infrastructure.InitDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = infrastructure.CloseDB() }()
	utils.LogInfo("Database initialized successfully")

	// Auto Migrate Entities
	if err := infrastructure.DB.AutoMigrate(
		&terminal_entities.Terminal{},
		&device_entities.Device{},
		&scene_entities.Scene{},
		&recordings_entities.Recording{},
		&common_entities.ScheduledNotification{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate entities: %w", err)
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
		defer func() { _ = badgerService.Close() }()
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
	deviceRepo := device_repositories.NewDeviceRepository(badgerService)
	terminalRepo := terminal_repositories.NewTerminalRepository(badgerService)

	// Initialize Modules
	commonModule := common.NewCommonModule(badgerService, vectorService, mqttService, terminalRepo, utils.GetConfig())
	tuyaModule := tuya.NewTuyaModule(badgerService, vectorService, deviceRepo, terminalRepo)
	mailModule := mail.NewMailModule(utils.GetConfig(), badgerService)

	terminalModule := terminal.NewTerminalModule(badgerService, deviceRepo, tuyaModule.AuthUseCase, tuyaModule.GetDeviceByIDUseCase, tuyaModule.DeviceControlUseCase)
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

	// 3. Terminal Routes (CRUD)
	terminalModule.RegisterRoutes(router, protected)

	// 3a. Mail Routes
	mailModule.RegisterRoutes(protected)

	// 4. Recordings Module
	recordingsModule := recordings.NewRecordingsModule(badgerService)
	recordingsModule.RegisterRoutes(router, protected)

	// 5. Speech & RAG Modules (migrated from stt-service)
	scfg := utils.GetConfig()
	// Log current configuration for diagnostic purposes
	utils.LogInfo("Startup: LLM Provider is set to '%s'", scfg.LLMProvider)
	utils.LogInfo("Startup: Application log level is '%s'", utils.GetCurrentLogLevelName())
	missing := []string{}
	// Validate mandatory config based on provider
	switch scfg.LLMProvider {
	case "gemini":
		if scfg.GeminiApiKey == "" {
			missing = append(missing, "GEMINI_API_KEY")
		}
	case "openai":
		if scfg.OpenAIApiKey == "" {
			missing = append(missing, "OPENAI_API_KEY")
		}
	case "groq":
		if scfg.GroqApiKey == "" {
			missing = append(missing, "GROQ_API_KEY")
		}
	case "orion":
		if scfg.OrionApiKey == "" {
			missing = append(missing, "ORION_API_KEY")
		}
		if scfg.OrionWhisperBaseURL == "" && scfg.WhisperLocalModel == "" {
			missing = append(missing, "ORION_WHISPER_BASE_URL (required for Orion Whisper)")
		}
	default:
		// If no provider or invalid, still check for common fallback
		if scfg.GeminiApiKey == "" && scfg.OrionApiKey == "" && scfg.OpenAIApiKey == "" && scfg.GroqApiKey == "" {
			missing = append(missing, "LLM API Key (GEMINI_API_KEY, OPENAI_API_KEY, GROQ_API_KEY, or ORION_API_KEY)")
		}
	}

	// Whisper check (only if not using multimodal/API-based providers that handle it)
	// Actually all our new providers handle it, but we might want local fallback.
	// For now, if provider is set, we trust its InitModule to fail if specific whisper model is missing.
	if scfg.MaxFileSize == 0 {
		missing = append(missing, "MAX_FILE_SIZE_MB")
	}
	if scfg.Port == "" {
		missing = append(missing, "PORT")
	}
	if len(missing) > 0 {
		return fmt.Errorf("speech/RAG config incomplete: %v", missing)
	}

	// 5. Models Module (Consolidated Whisper, RAG, and Pipeline)
	// This replaces the direct Go RAG and Speech routes
	models.InitModule(
		protected,
		scfg,
		badgerService,
		vectorService,
		tuyaModule.AuthUseCase,
		tuyaModule.DeviceControlUseCase,
		mqttService,
		terminalRepo,
		recordingsModule.SaveRecordingUseCase,
		commonModule.StorageProvider,
	)

	// 5b. Models-v1 Module (v1 routes: /api/models/v1/...)
	// This provides access to Python AI services via gRPC/REST
	models_v1.InitModule(protected, scfg)

	// 6. Scene Module
	sceneModule := scene.NewSceneModule(infrastructure.DB, tuyaModule.DeviceControlUseCase, mqttService)
	sceneModule.RegisterRoutes(protected)

	// Register Health at the end so it appears last in Swagger
	router.GET("/api/health", commonModule.HealthController.CheckHealth)

	// Start notification scheduler worker for WA notifications
	notificationWorker := services.NewNotificationSchedulerWorker(scfg.WANotificationBaseURL)
	notificationWorker.Start()
	utils.LogInfo("Notification scheduler worker started with WA endpoint: %s", scfg.WANotificationBaseURL)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		utils.LogInfo("Shutting down server...")
		notificationWorker.Stop()
		os.Exit(0)
	}()

	port := scfg.Port
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  1 * time.Hour, // Long timeout for slow uploads
		WriteTimeout: 1 * time.Hour, // Long timeout for slow responses
		IdleTimeout:  5 * time.Minute,
	}

	utils.LogInfo("Server starting on :%s", port)
	return server.ListenAndServe()
}
