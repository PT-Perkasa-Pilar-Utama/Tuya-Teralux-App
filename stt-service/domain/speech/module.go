package speech

import (
	"stt-service/domain/common/config"
	"stt-service/domain/speech/controllers"
	"stt-service/domain/speech/repositories"
	"stt-service/domain/speech/routes"
	"stt-service/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

func InitModule(router *gin.Engine, cfg *config.Config) {
	// Repositories
	whisperRepo := repositories.NewWhisperRepository()
	ollamaRepo := repositories.NewOllamaRepository()

	// Usecases
	transcriptionUsecase := usecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo, cfg)

	// Controllers
	transcriptionController := controllers.NewTranscriptionController(transcriptionUsecase, cfg)

	// Routes
	routes.RegisterRoutes(router, transcriptionController)
}
