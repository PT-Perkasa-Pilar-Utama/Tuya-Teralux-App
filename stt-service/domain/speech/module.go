package speech

import (
	"stt-service/domain/speech/controllers"
	"stt-service/domain/speech/repositories"
	"stt-service/domain/speech/routes"
	"stt-service/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

func InitModule(router *gin.Engine) {
	// Repositories
	whisperRepo := repositories.NewWhisperRepository()
	ollamaRepo := repositories.NewOllamaRepository()

	// Usecases
	transcriptionUsecase := usecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo)

	// Controllers
	transcriptionController := controllers.NewTranscriptionController(transcriptionUsecase)

	// Routes
	routes.RegisterRoutes(router, transcriptionController)
}
