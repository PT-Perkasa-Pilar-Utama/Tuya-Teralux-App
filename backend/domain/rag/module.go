package rag

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/controllers"
	"teralux_app/domain/rag/routes"
	"teralux_app/domain/rag/usecases"
	speechRepos "teralux_app/domain/speech/repositories"

	"github.com/gin-gonic/gin"
)

func InitModule(r *gin.Engine, cfg *utils.Config) {
	// Initialize Dependencies
	vectorSvc := infrastructure.NewVectorService()
	ollamaRepo := speechRepos.NewOllamaRepository()
	ragUsecase := usecases.NewRAGUsecase(vectorSvc, ollamaRepo, cfg)
	ragController := controllers.NewRAGController(ragUsecase)

	// Setup Routes
	routes.SetupRAGRoutes(r, ragController)
}
