package speech

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	"teralux_app/domain/common/services"
	speechControllers "teralux_app/domain/speech/controllers"
	speechRoutes "teralux_app/domain/speech/routes"
	speechUsecases "teralux_app/domain/speech/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Speech module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badgerSvc *infrastructure.BadgerService, ragRefineUC usecases.RefineUseCase, tuyaAuthUseCase tuyaUsecases.TuyaAuthUseCase, mqttSvc *infrastructure.MqttService, saveRecordingUseCase recordingUsecases.SaveRecordingUseCase) {
	// Services
	geminiService := services.NewGeminiService(cfg)
	localService := services.NewWhisperLocalService(cfg)

	// Usecases
	shortCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "transcribe:task:")
	longCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "transcribe_long:task:")
	whisperCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "whisper:task:")
	whisperProxyUsecase := speechUsecases.NewWhisperProxyUsecase(whisperCache, cfg)

	// Setup Whisper Clients
	// whisperProxyUsecase now implements WhisperClient directly.

	// Select Whisper Client based on configuration
	var whisperClient speechUsecases.WhisperClient
	
	
	if cfg.LLMProvider == "gemini" {
		utils.LogInfo("Speech: Using Gemini Whisper (Multimodal)")
		whisperClient = geminiService
	} else if cfg.LLMProvider == "orion" {
		if cfg.OrionWhisperBaseURL != "" {
			utils.LogInfo("Speech: Using Remote Whisper (PPU/Orion)")
			whisperClient = whisperProxyUsecase
		} else {
			utils.LogFatal("Speech: LLM_PROVIDER is 'orion' but ORION_WHISPER_BASE_URL is not set.")
		}
	} else {
		utils.LogFatal("Speech: Invalid or missing LLM_PROVIDER. Set it to 'gemini' or 'orion'.")
	}

	// Feature Usecases (1 Route 1 Usecase)
	transcribeUC := speechUsecases.NewTranscribeUseCase(whisperClient, ragRefineUC, shortCache, cfg, mqttSvc)
	transcribeWhisperCppUC := speechUsecases.NewTranscribeWhisperCppUseCase(localService, ragRefineUC, longCache, cfg)
	getStatusUC := speechUsecases.NewGetTranscriptionStatusUseCase(shortCache, longCache, whisperProxyUsecase)

	// Controllers
	transcribeController := speechControllers.NewSpeechTranscribeController(transcribeUC, saveRecordingUseCase, cfg, mqttSvc)
	transcribeController.StartMqttSubscription()
	statusController := speechControllers.NewSpeechTranscribeStatusController(getStatusUC)
	whisperCppController := speechControllers.NewSpeechTranscribeWhisperCppController(transcribeWhisperCppUC, saveRecordingUseCase, cfg)
	ppuController := speechControllers.NewSpeechTranscribePPUController(whisperProxyUsecase, saveRecordingUseCase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(protected, transcribeController, statusController, whisperCppController, ppuController)
}
