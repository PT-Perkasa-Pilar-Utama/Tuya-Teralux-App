package whisper

import (
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	ragUsecases "sensio/domain/models/rag/usecases"
	recordingUsecases "sensio/domain/recordings/usecases"
	whisperControllers "sensio/domain/models/whisper/controllers"
	whisperDtos "sensio/domain/models/whisper/dtos"
	whisperRoutes "sensio/domain/models/whisper/routes"
	whisperUsecases "sensio/domain/models/whisper/usecases"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"time"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Whisper module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badgerSvc *infrastructure.BadgerService, ragRefineUC ragUsecases.RefineUseCase, tuyaAuthUseCase tuyaUsecases.TuyaAuthUseCase, mqttSvc *infrastructure.MqttService, saveRecordingUseCase recordingUsecases.SaveRecordingUseCase) (whisperUsecases.TranscribeUseCase, whisperUsecases.UploadSessionUseCase) {
	// Services
	geminiService := services.NewGeminiService(cfg)
	localService := services.NewWhisperLocalService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	groqService := services.NewGroqService(cfg)
	orionService := services.NewOrionService(cfg)

	// Usecases
	shortCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "cache:transcribe:task:")

	// Select Whisper Client based on configuration
	var whisperClient whisperUsecases.WhisperClient

	switch cfg.LLMProvider {
	case "gemini":
		utils.LogInfo("Whisper: Using Gemini Whisper (Multimodal)")
		whisperClient = geminiService
	case "openai":
		utils.LogInfo("Whisper: Using OpenAI Whisper")
		whisperClient = openaiService
	case "groq":
		utils.LogInfo("Whisper: Using Groq Whisper")
		whisperClient = groqService
	case "orion":
		if cfg.OrionWhisperBaseURL != "" {
			utils.LogInfo("Whisper: Using Remote Whisper (Orion)")
			whisperClient = orionService
		} else {
			utils.LogFatal("Whisper: LLM_PROVIDER is 'orion' but ORION_WHISPER_BASE_URL is not set.")
		}
	default:
		utils.LogFatal("Whisper: Invalid or missing LLM_PROVIDER. Set it to 'gemini', 'openai', 'groq', or 'orion'.")
	}

	shortCacheStore := tasks.NewStatusStore[whisperDtos.AsyncTranscriptionStatusDTO]()

	// Feature Usecases (1 Route 1 Usecase)
	transcribeUC := whisperUsecases.NewTranscribeUseCase(whisperClient, localService, ragRefineUC, shortCacheStore, shortCache, cfg, mqttSvc)

	// Models Usecases (Async)
	geminiModelUC := whisperUsecases.NewTranscribeGeminiModelUseCase(geminiService, shortCacheStore, shortCache, cfg)
	openaiModelUC := whisperUsecases.NewTranscribeOpenAIModelUseCase(openaiService, shortCacheStore, shortCache, cfg)
	groqModelUC := whisperUsecases.NewTranscribeGroqModelUseCase(groqService, shortCacheStore, shortCache, cfg)
	orionModelUC := whisperUsecases.NewTranscribeOrionModelUseCase(orionService, shortCacheStore, shortCache, cfg)
	cppModelUC := whisperUsecases.NewTranscribeWhisperCppModelUseCase(localService, shortCacheStore, shortCache, cfg)
	uploadSessionUC := whisperUsecases.NewUploadSessionUseCase(badgerSvc, cfg)

	// Status Usecase (Generic)
	getStatusUC := tasks.NewGenericStatusUseCase(shortCache, shortCacheStore)

	// Phase 4: Start background cleanup worker for upload sessions
	if cfg.EnableChunkUpload {
		cleanupInterval, err := time.ParseDuration(cfg.ChunkUploadCleanupInterval)
		if err != nil {
			cleanupInterval = 10 * time.Minute
		}
		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for {
				select {
				case now := <-ticker.C:
					count, err := uploadSessionUC.CleanupExpiredSessions(now)
					if err != nil {
						utils.LogError("Whisper: Upload session cleanup failed: %v", err)
					} else if count > 0 {
						utils.LogInfo("Whisper: Cleaned up %d expired upload sessions", count)
					}
				}
			}
		}()
	}

	// Controllers
	transcribeController := whisperControllers.NewWhisperTranscribeController(transcribeUC, saveRecordingUseCase, uploadSessionUC, cfg, mqttSvc)
	if err := transcribeController.StartMqttSubscription(); err != nil {
		utils.LogError("Whisper module MQTT subscription failed: %v", err)
	}
	statusController := whisperControllers.NewWhisperTranscribeStatusController(getStatusUC)
	uploadSessionController := whisperControllers.NewUploadSessionController(uploadSessionUC, transcribeUC)

	geminiController := whisperControllers.NewWhisperModelsGeminiController(geminiModelUC, saveRecordingUseCase, cfg)
	openaiController := whisperControllers.NewWhisperModelsOpenAIController(openaiModelUC, saveRecordingUseCase, cfg)
	groqController := whisperControllers.NewWhisperModelsGroqController(groqModelUC, saveRecordingUseCase, cfg)
	orionModelController := whisperControllers.NewWhisperModelsOrionController(orionModelUC, saveRecordingUseCase, cfg)
	cppModelController := whisperControllers.NewWhisperModelsWhisperCppController(cppModelUC, saveRecordingUseCase, cfg)

	// Routes
	whisperRoutes.SetupWhisperRoutes(
		protected,
		transcribeController,
		statusController,
		geminiController,
		openaiController,
		groqController,
		orionModelController,
		cppModelController,
		uploadSessionController,
	)

	return transcribeUC, uploadSessionUC
}
