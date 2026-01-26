package speech

import (
	"teralux_app/domain/common/config"
	"teralux_app/domain/speech/controllers"
	"teralux_app/domain/speech/repositories"
	"teralux_app/domain/speech/routes"
	"teralux_app/domain/speech/usecases"

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
