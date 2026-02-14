package speech

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	speechControllers "teralux_app/domain/speech/controllers"
	"teralux_app/domain/speech/repositories"
	speechRoutes "teralux_app/domain/speech/routes"
	speechUsecases "teralux_app/domain/speech/usecases"
	"teralux_app/domain/speech/utilities"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Speech module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badgerSvc *infrastructure.BadgerService, ragRefineUC usecases.RefineUseCase, tuyaAuthUseCase tuyaUsecases.TuyaAuthUseCase, mqttSvc *infrastructure.MqttService, saveRecordingUseCase recordingUsecases.SaveRecordingUseCase) {
	// Repositories
	whisperCppRepo := repositories.NewWhisperCppRepository(cfg)
	whisperOrionRepo := repositories.NewWhisperOrionRepository(cfg)

	// Usecases
	shortCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "transcribe:task:")
	longCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "transcribe_long:task:")
	whisperCache := tasks.NewBadgerTaskCacheFromService(badgerSvc, "whisper:task:")
	whisperProxyUsecase := speechUsecases.NewWhisperProxyUsecase(whisperCache, cfg)

	// Setup Whisper Clients with adapters
	ppuClient := utilities.NewPPUWhisperClient(whisperProxyUsecase)
	orionClient := utilities.NewOrionWhisperClient(whisperOrionRepo)
	localClient := utilities.NewLocalWhisperClient(whisperCppRepo, cfg.WhisperModelPath)

	// Whisper client with automatic fallback (PPU -> Orion -> Local)
	whisperClientWithFallback := utilities.NewWhisperClientWithFallback(ppuClient, orionClient, localClient)

	// Feature Usecases (1 Route 1 Usecase)
	transcribeUC := speechUsecases.NewTranscribeUseCase(whisperClientWithFallback, ragRefineUC, shortCache, cfg)
	transcribeWhisperCppUC := speechUsecases.NewTranscribeWhisperCppUseCase(whisperCppRepo, ragRefineUC, longCache, cfg)
	getStatusUC := speechUsecases.NewGetTranscriptionStatusUseCase(shortCache, longCache, whisperProxyUsecase)

	// Controllers
	transcribeController := speechControllers.NewSpeechTranscribeController(transcribeUC, saveRecordingUseCase, cfg)
	statusController := speechControllers.NewSpeechTranscribeStatusController(getStatusUC)
	whisperCppController := speechControllers.NewSpeechTranscribeWhisperCppController(transcribeWhisperCppUC, saveRecordingUseCase, cfg)
	ppuController := speechControllers.NewSpeechTranscribePPUController(whisperProxyUsecase, saveRecordingUseCase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(protected, transcribeController, statusController, whisperCppController, ppuController)
}
