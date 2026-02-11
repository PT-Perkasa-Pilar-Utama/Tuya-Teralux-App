package speech

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	speechControllers "teralux_app/domain/speech/controllers"
	"teralux_app/domain/speech/repositories"
	speechRoutes "teralux_app/domain/speech/routes"
	speechUsecases "teralux_app/domain/speech/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Speech module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badgerSvc *infrastructure.BadgerService, ragUsecase *usecases.RAGUsecase, tuyaAuthUseCase *tuyaUsecases.TuyaAuthUseCase, mqttSvc *infrastructure.MqttService, saveRecordingUseCase recordingUsecases.SaveRecordingUseCase) {
	// Repositories
	whisperRepo := repositories.NewWhisperRepository(cfg)
	taskRepo := repositories.NewTranscriptionTaskRepository(badgerSvc)

	// Usecases
	whisperProxyUsecase := speechUsecases.NewWhisperProxyUsecase(badgerSvc, cfg)
	
	// Feature Usecases (1 Route 1 Usecase)
	transcribeUC := speechUsecases.NewTranscribeUseCase(whisperRepo, whisperProxyUsecase, ragUsecase, taskRepo, cfg)
	transcribeWhisperCppUC := speechUsecases.NewTranscribeWhisperCppUseCase(whisperRepo, ragUsecase, taskRepo, cfg)
	getStatusUC := speechUsecases.NewGetTranscriptionStatusUseCase(taskRepo, whisperProxyUsecase)

	// Controllers
	transcriptionController := speechControllers.NewTranscriptionController(transcribeUC, transcribeWhisperCppUC, getStatusUC, cfg)
	whisperProxyController := speechControllers.NewWhisperProxyController(whisperProxyUsecase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(protected, transcriptionController, whisperProxyController)
}
