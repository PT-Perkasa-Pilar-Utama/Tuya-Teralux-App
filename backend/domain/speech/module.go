package speech

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	speechControllers "teralux_app/domain/speech/controllers"
	"teralux_app/domain/speech/repositories"
	speechRoutes "teralux_app/domain/speech/routes"
	speechUsecases "teralux_app/domain/speech/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Speech module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badgerSvc *infrastructure.BadgerService, ragUsecase *usecases.RAGUsecase, tuyaAuthUseCase *tuyaUsecases.TuyaAuthUseCase, mqttSvc *infrastructure.MqttService) {
	// Repositories
	whisperRepo := repositories.NewWhisperRepository(cfg)

	// Usecases
	whisperProxyUsecase := speechUsecases.NewWhisperProxyUsecase(badgerSvc, cfg)
	transcriptionUsecase := speechUsecases.NewTranscriptionUsecase(whisperRepo, cfg, ragUsecase, tuyaAuthUseCase, whisperProxyUsecase, badgerSvc)


	// Controllers
	transcriptionController := speechControllers.NewTranscriptionController(transcriptionUsecase, whisperProxyUsecase, cfg)
	whisperProxyController := speechControllers.NewWhisperProxyController(whisperProxyUsecase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(protected, transcriptionController, whisperProxyController)
}
