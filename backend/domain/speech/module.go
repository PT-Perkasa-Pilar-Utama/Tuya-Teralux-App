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
	ollamaRepo := repositories.NewOllamaRepository()
	geminiRepo := repositories.NewGeminiRepository()
	antigravityRepo := repositories.NewAntigravityRepository()
	mqttRepo := repositories.NewMqttRepository(mqttSvc, cfg)

	// Usecases
	transcriptionUsecase := speechUsecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo, geminiRepo, antigravityRepo, mqttRepo, cfg, ragUsecase, tuyaAuthUseCase, badgerSvc)
	whisperProxyUsecase := speechUsecases.NewWhisperProxyUsecase(badgerSvc, cfg, mqttRepo)

	// Start MQTT Listener
	transcriptionUsecase.StartListening()

	// Controllers
	transcriptionController := speechControllers.NewTranscriptionController(transcriptionUsecase, whisperProxyUsecase, cfg)
	whisperProxyController := speechControllers.NewWhisperProxyController(whisperProxyUsecase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(protected, transcriptionController, whisperProxyController)
}
