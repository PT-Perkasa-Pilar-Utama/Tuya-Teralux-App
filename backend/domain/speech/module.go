package speech

import (
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
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, ragUsecase *usecases.RAGUsecase, tuyaAuthUseCase *tuyaUsecases.TuyaAuthUseCase) {
	// Repositories
	whisperRepo := repositories.NewWhisperRepository(cfg)
	ollamaRepo := repositories.NewOllamaRepository()
	geminiRepo := repositories.NewGeminiRepository()
	antigravityRepo := repositories.NewAntigravityRepository()
	mqttRepo := repositories.NewMqttRepository(cfg)

	// Usecases
	transcriptionUsecase := speechUsecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo, geminiRepo, antigravityRepo, mqttRepo, cfg, ragUsecase, tuyaAuthUseCase)

	// Start MQTT Listener
	transcriptionUsecase.StartListening()

	// Controllers
	transcriptionController := speechControllers.NewTranscriptionController(transcriptionUsecase, cfg)

	// Routes
	speechRoutes.SetupSpeechRoutes(protected, transcriptionController)
}
