package rag

import (
	"stt-service/domain/rag/controllers"
	"stt-service/domain/rag/repositories"
	"stt-service/domain/rag/routes"
	"stt-service/domain/rag/usecases"

	"github.com/gin-gonic/gin"
)

func InitModule(r *gin.Engine) {
	// Initialize Dependencies
	vectorRepo := repositories.NewVectorRepository()
	ragUsecase := usecases.NewRAGUsecase(vectorRepo)
	ragController := controllers.NewRAGController(ragUsecase)

	// Setup Routes
	routes.SetupRAGRoutes(r, ragController)
}
