package speech

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/controllers"
	"teralux_app/domain/speech/repositories"
	"teralux_app/domain/speech/routes"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the Speech module with the protected router group.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config) {
	// Repositories
	// Repositories
	whisperRepo := repositories.NewWhisperRepository()
	ollamaRepo := repositories.NewOllamaRepository()
	mqttRepo := repositories.NewMqttRepository(cfg)

	// Usecases
	transcriptionUsecase := usecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo, mqttRepo, cfg)

	// Start MQTT Listener
	transcriptionUsecase.StartListening()

	// Controllers
	transcriptionController := controllers.NewTranscriptionController(transcriptionUsecase, cfg)

	// Routes
	routes.SetupSpeechRoutes(protected, transcriptionController)
}
