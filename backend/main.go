package main

// Trigger documentation refresh build

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"sensio/domain/common"
	"sensio/domain/common/controllers"
	"sensio/domain/common/interfaces"
	"sensio/domain/common/middlewares"
	"sensio/domain/common/utils"
	"sensio/domain/download_token"
	"sensio/domain/infrastructure"
	"sensio/domain/mail"
	"sensio/domain/models"
	notification_entities "sensio/domain/notification/entities"
	"sensio/domain/notification"
	notification_services "sensio/domain/notification/services"
	recordings "sensio/domain/recordings"
	recordings_entities "sensio/domain/recordings/entities"
	"sensio/domain/scene"
	scene_entities "sensio/domain/scene/entities"
	"sensio/domain/terminal"
	device_entities "sensio/domain/terminal/device/entities"
	device_repositories "sensio/domain/terminal/device/repositories"
	device_status_entities "sensio/domain/terminal/device_status/entities"
	terminal_entities "sensio/domain/terminal/terminal/entities"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	"sensio/domain/tuya"
)

type terminalRepoAdapter struct {
	repo *terminal_repositories.TerminalRepository
}

func (a *terminalRepoAdapter) toEntity(t *interfaces.Terminal) *terminal_entities.Terminal {
	var aiProvider, aiEngineProfile *string
	if t.AiProvider != "" {
		s := t.AiProvider
		aiProvider = &s
	}
	if t.AiEngineProfile != "" {
		s := t.AiEngineProfile
		aiEngineProfile = &s
	}
	deviceTypeID := t.DeviceTypeID
	return &terminal_entities.Terminal{
		ID:              t.ID,
		MacAddress:      t.MacAddress,
		RoomID:          t.RoomID,
		TuyaUID:         t.TuyaUID,
		Name:            t.Name,
		DeviceTypeID:    fmt.Sprintf("%d", deviceTypeID),
		AiProvider:      aiProvider,
		AiEngineProfile: aiEngineProfile,
	}
}

func (a *terminalRepoAdapter) fromEntity(t *terminal_entities.Terminal) interfaces.Terminal {
	var aiProvider, aiEngineProfile string
	if t.AiProvider != nil {
		aiProvider = *t.AiProvider
	}
	if t.AiEngineProfile != nil {
		aiEngineProfile = *t.AiEngineProfile
	}
	deviceTypeID, _ := strconv.Atoi(t.DeviceTypeID)
	return interfaces.Terminal{
		ID:              t.ID,
		MacAddress:      t.MacAddress,
		RoomID:          t.RoomID,
		TuyaUID:         t.TuyaUID,
		Name:            t.Name,
		DeviceTypeID:    deviceTypeID,
		AiProvider:      aiProvider,
		AiEngineProfile: aiEngineProfile,
	}
}

func (a *terminalRepoAdapter) Create(ctx context.Context, terminal *interfaces.Terminal) error {
	return a.repo.Create(a.toEntity(terminal))
}

func (a *terminalRepoAdapter) GetAll(ctx context.Context) ([]interfaces.Terminal, error) {
	terms, err := a.repo.GetAll()
	if err != nil {
		return nil, err
	}
	result := make([]interfaces.Terminal, len(terms))
	for i, t := range terms {
		result[i] = a.fromEntity(&t)
	}
	return result, nil
}

func (a *terminalRepoAdapter) GetAllPaginated(ctx context.Context, offset, limit int, roomID *string) ([]interfaces.Terminal, int64, error) {
	terms, total, err := a.repo.GetAllPaginated(offset, limit, roomID)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interfaces.Terminal, len(terms))
	for i, t := range terms {
		result[i] = a.fromEntity(&t)
	}
	return result, total, nil
}

func (a *terminalRepoAdapter) GetByID(ctx context.Context, id string) (*interfaces.Terminal, error) {
	t, err := a.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	v := a.fromEntity(t)
	return &v, nil
}

func (a *terminalRepoAdapter) GetByMacAddress(ctx context.Context, macAddress string) (*interfaces.Terminal, error) {
	t, err := a.repo.GetByMacAddress(macAddress)
	if err != nil {
		return nil, err
	}
	v := a.fromEntity(t)
	return &v, nil
}

func (a *terminalRepoAdapter) GetByRoomID(ctx context.Context, roomID string) ([]interfaces.Terminal, error) {
	terms, err := a.repo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}
	result := make([]interfaces.Terminal, len(terms))
	for i, t := range terms {
		result[i] = a.fromEntity(&t)
	}
	return result, nil
}

func (a *terminalRepoAdapter) Update(ctx context.Context, terminal *interfaces.Terminal) error {
	return a.repo.Update(a.toEntity(terminal))
}

func (a *terminalRepoAdapter) Delete(ctx context.Context, id string) error {
	return a.repo.Delete(id)
}

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

// @tag.name 05. Recordings
// @tag.description Recordings management endpoints

// @tag.name 06. Mail
// @tag.description Mail service endpoints

// @tag.name 07. Common
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
		&device_status_entities.DeviceStatus{},
		&scene_entities.Scene{},
		&recordings_entities.Recording{},
		&recordings_entities.AudioUploadStatus{},
		&notification_entities.ScheduledNotification{},
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

	terminalRepoAdapter := &terminalRepoAdapter{repo: terminalRepo}

	// Initialize Modules
	tuyaModule := tuya.NewTuyaModule(badgerService, vectorService, deviceRepo, terminalRepo)

	var downloadTokenCreator interfaces.DownloadTokenCreator
	storageProvider, _ := infrastructure.NewStorageProvider(utils.GetConfig())
	downloadTokenCreator = download_token.NewDownloadTokenService(storageProvider)

	commonModule := common.NewCommonModule(badgerService, vectorService, mqttService, terminalRepoAdapter, utils.GetConfig(), tuyaModule.AuthUseCase, downloadTokenCreator)
	mailModule := mail.NewMailModule(utils.GetConfig(), badgerService)

	notificationModule := notification.NewNotificationModule(badgerService, mqttService, terminalRepo)

	terminalModule := terminal.NewTerminalModule(badgerService, deviceRepo, tuyaModule.AuthUseCase, tuyaModule.GetDeviceByIDUseCase, tuyaModule.DeviceControlUseCase)
	// Register Routes
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware(tuyaModule.AuthUseCase))
	protected.Use(middlewares.TuyaErrorMiddleware())

	// Static File Serving (protected via auth middleware)
	protected.GET("/uploads/:filename", middlewares.AuthMiddleware(tuyaModule.AuthUseCase), controllers.ServeProtectedUploads())

	// 1. Common Routes (Health, Cache)
	commonModule.RegisterRoutes(router, protected)

	// 2. Tuya Routes (Auth, Device Control)
	tuyaModule.RegisterRoutes(router, protected)

	// 3. Terminal Routes (CRUD)
	terminalModule.RegisterRoutes(router, protected)

	// 3a. Mail Routes
	mailModule.RegisterRoutes(protected)

	// 3b. Notification Routes
	notificationModule.RegisterRoutes(protected)

	// 4. Recordings Module
	recordingsModule := recordings.NewRecordingsModule(badgerService, commonModule.StorageProvider)
	recordingsModule.RegisterRoutes(router, protected)

	download_tokenModule := download_token.NewDownloadTokenModule(commonModule.StorageProvider)
	download_tokenModule.RegisterRoutes(protected)

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
	models.InitModule(
		protected,
		scfg,
		badgerService,
		vectorService,
		tuyaModule.AuthUseCase,
		tuyaModule.DeviceControlUseCase,
		mqttService,
		terminalRepoAdapter,
		recordingsModule.SaveRecordingUseCase,
		commonModule.StorageProvider,
		downloadTokenCreator,
	)

	// 6. Scene Module
	sceneModule := scene.NewSceneModule(infrastructure.DB, tuyaModule.DeviceControlUseCase, mqttService)
	sceneModule.RegisterRoutes(protected)

	// Register Health at the end so it appears last in Swagger
	router.GET("/api/health", commonModule.HealthController.CheckHealth)

	// Start notification scheduler worker for WA notifications
	notificationWorker := notification_services.NewNotificationSchedulerWorker(scfg.WANotificationBaseURL)
	notificationWorker.Start()
	utils.LogInfo("Notification scheduler worker started with WA endpoint: %s", scfg.WANotificationBaseURL)

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

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		utils.LogInfo("Shutting down server...")
		notificationWorker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	utils.LogInfo("Server starting on :%s", port)
	return server.ListenAndServe()
}
