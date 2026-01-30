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

// InitModule initializes RAG module with protected router group, configuration and optional persistence.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService) *usecases.RAGUsecase {
	// Initialize Dependencies
	var llmRepo usecases.LLMClient
	if cfg.LLMProvider == "gemini" {
		llmRepo = speechRepos.NewGeminiRepository()
	} else {
		llmRepo = speechRepos.NewOllamaRepository()
	}

	ragUsecase := usecases.NewRAGUsecase(vectorSvc, llmRepo, cfg, badger)
	ragController := controllers.NewRAGController(ragUsecase)

	// Setup Routes
	routes.SetupRAGRoutes(protected, ragController)

	return ragUsecase
}
