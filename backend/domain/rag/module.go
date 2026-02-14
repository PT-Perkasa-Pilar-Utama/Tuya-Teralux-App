package rag

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/controllers"
	ragdtos "teralux_app/domain/rag/dtos"
	ragRepos "teralux_app/domain/rag/repositories"
	"teralux_app/domain/rag/routes"
	"teralux_app/domain/rag/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes RAG module with protected router group, configuration and optional persistence.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, tuyaAuth tuyaUsecases.TuyaAuthUseCase) usecases.RefineUseCase {
	// Initialize Dependencies
	geminiRepo := ragRepos.NewGeminiRepository()
	orionRepo := ragRepos.NewOrionRepository()

	// Use Fallback client: Orion (Primary) -> Gemini (Fallback)
	llmClient := usecases.NewLLMClientFallback(orionRepo, geminiRepo)

	// Initialize Shared Store
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()
	cache := tasks.NewBadgerTaskCache(badger, "rag:task:")

	// Initialize Usecases
	refineUC := usecases.NewRefineUseCase(llmClient, cfg)
	translateUC := usecases.NewTranslateUseCase(llmClient, cfg, cache, store)
	summaryUC := usecases.NewSummaryUseCase(llmClient, cfg, cache, store)
	statusUC := usecases.NewRAGStatusUseCase(cache, store)
	controlUC := usecases.NewControlUseCase(vectorSvc, llmClient, cfg, cache, tuyaAuth, store)

	// Setup Routes
	routes.SetupRAGRoutes(
		protected,
		controllers.NewRAGControlController(controlUC, statusUC, cfg),
		controllers.NewRAGTranslateController(translateUC),
		controllers.NewRAGSummaryController(summaryUC),
		controllers.NewRAGStatusController(statusUC),
	)

	return refineUC
}
