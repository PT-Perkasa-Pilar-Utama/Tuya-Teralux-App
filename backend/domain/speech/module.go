package speech

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/services"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	speechControllers "teralux_app/domain/speech/controllers"
	speechdtos "teralux_app/domain/speech/dtos"
	speechRoutes "teralux_app/domain/speech/routes"
	speechUsecases "teralux_app/domain/speech/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Speech module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badgerSvc *infrastructure.BadgerService, ragRefineUC ragUsecases.RefineUseCase, tuyaAuthUseCase tuyaUsecases.TuyaAuthUseCase, mqttSvc *infrastructure.MqttService, saveRecordingUseCase recordingUsecases.SaveRecordingUseCase) {
	// Services
	geminiService := services.NewGeminiService(cfg)
	localService := services.NewWhisperLocalService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	groqService := services.NewGroqService(cfg)
	orionService := services.NewOrionService(cfg)

	// Usecases
	shortCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "transcribe:task:")

	// Select Whisper Client based on configuration
	var whisperClient speechUsecases.WhisperClient

	switch cfg.LLMProvider {
	case "gemini":
		utils.LogInfo("Speech: Using Gemini Whisper (Multimodal)")
		whisperClient = geminiService
	case "openai":
		utils.LogInfo("Speech: Using OpenAI Whisper")
		whisperClient = openaiService
	case "groq":
		utils.LogInfo("Speech: Using Groq Whisper")
		whisperClient = groqService
	case "orion":
		if cfg.OrionWhisperBaseURL != "" {
			utils.LogInfo("Speech: Using Remote Whisper (Orion)")
			whisperClient = orionService
		} else {
			utils.LogFatal("Speech: LLM_PROVIDER is 'orion' but ORION_WHISPER_BASE_URL is not set.")
		}
	default:
		utils.LogFatal("Speech: Invalid or missing LLM_PROVIDER. Set it to 'gemini', 'openai', 'groq', or 'orion'.")
	}

	shortCacheStore := tasks.NewStatusStore[speechdtos.AsyncTranscriptionStatusDTO]()
	
	// Feature Usecases (1 Route 1 Usecase)
	transcribeUC := speechUsecases.NewTranscribeUseCase(whisperClient, localService, ragRefineUC, shortCacheStore, shortCache, cfg, mqttSvc)

	// Models Usecases (Async)
	geminiModelUC := speechUsecases.NewTranscribeGeminiModelUseCase(geminiService, shortCacheStore, shortCache, cfg)
	openaiModelUC := speechUsecases.NewTranscribeOpenAIModelUseCase(openaiService, shortCacheStore, shortCache, cfg)
	groqModelUC := speechUsecases.NewTranscribeGroqModelUseCase(groqService, shortCacheStore, shortCache, cfg)
	orionModelUC := speechUsecases.NewTranscribeOrionModelUseCase(orionService, shortCacheStore, shortCache, cfg)
	cppModelUC := speechUsecases.NewTranscribeWhisperCppModelUseCase(localService, shortCacheStore, shortCache, cfg)

	// Status Usecase (Generic)
	getStatusUC := tasks.NewGenericStatusUseCase(shortCache, shortCacheStore)

	// Controllers
	transcribeController := speechControllers.NewSpeechTranscribeController(transcribeUC, saveRecordingUseCase, cfg, mqttSvc)
	transcribeController.StartMqttSubscription()
	statusController := speechControllers.NewSpeechTranscribeStatusController(getStatusUC)

	geminiController := speechControllers.NewSpeechModelsGeminiController(geminiModelUC, saveRecordingUseCase, cfg)
	openaiController := speechControllers.NewSpeechModelsOpenAIController(openaiModelUC, saveRecordingUseCase, cfg)
	groqController := speechControllers.NewSpeechModelsGroqController(groqModelUC, saveRecordingUseCase, cfg)
	orionModelController := speechControllers.NewSpeechModelsOrionController(orionModelUC, saveRecordingUseCase, cfg)
	cppModelController := speechControllers.NewSpeechModelsWhisperCppController(cppModelUC, saveRecordingUseCase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(
		protected,
		transcribeController,
		statusController,
		geminiController,
		openaiController,
		groqController,
		orionModelController,
		cppModelController,
	)
}
