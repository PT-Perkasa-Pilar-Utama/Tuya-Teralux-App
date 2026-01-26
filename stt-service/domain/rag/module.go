package rag

import (
	"stt-service/domain/common/config"
	"stt-service/domain/rag/controllers"
	"stt-service/domain/rag/repositories"
	"stt-service/domain/rag/routes"
	"stt-service/domain/rag/usecases"
	speechRepos "stt-service/domain/speech/repositories"

	"github.com/gin-gonic/gin"
)

func InitModule(r *gin.Engine, cfg *config.Config) {
	// Initialize Dependencies
	vectorRepo := repositories.NewVectorRepository()
	ollamaRepo := speechRepos.NewOllamaRepository()
	ragUsecase := usecases.NewRAGUsecase(vectorRepo, ollamaRepo, cfg)
	ragController := controllers.NewRAGController(ragUsecase)

	// Setup Routes
	routes.SetupRAGRoutes(r, ragController)
}
