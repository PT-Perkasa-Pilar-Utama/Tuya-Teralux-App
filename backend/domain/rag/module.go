package rag

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/controllers"
	ragdtos "teralux_app/domain/rag/dtos"
	ragRepos "teralux_app/domain/rag/repositories"
	"teralux_app/domain/rag/routes"
	"teralux_app/domain/rag/services"
	"teralux_app/domain/rag/usecases"
	"teralux_app/domain/rag/utilities"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes RAG module with protected router group, configuration and optional persistence.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, tuyaAuth tuyaUsecases.TuyaAuthUseCase, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, mqttSvc *infrastructure.MqttService) usecases.RefineUseCase {
	// Initialize Dependencies
	orionRepo := ragRepos.NewOrionRepository()
	geminiRepo := ragRepos.NewGeminiRepository()
	ollamaRepo := ragRepos.NewOllamaRepository()

	// Use 3-level Fallback client: Orion (Primary) -> Gemini (Secondary) -> Ollama (Tertiary)
	llmClient := utilities.NewLLMClientWithFallback(orionRepo, geminiRepo, ollamaRepo)

	// Initialize Shared Store
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()
	cache := tasks.NewBadgerTaskCacheFromService(badger, "rag:task:")

	// Initialize Usecases
	refineUC := usecases.NewRefineUseCase(llmClient, cfg)
	translateUC := usecases.NewTranslateUseCase(llmClient, cfg, cache, store)
	pdfRenderer := services.NewMarotoSummaryPDFRenderer()
	summaryUC := usecases.NewSummaryUseCase(llmClient, cfg, cache, store, pdfRenderer)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	controlUC := usecases.NewControlUseCase(llmClient, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth)
	chatUC := usecases.NewChatUseCase(llmClient, cfg, badger, controlUC, translateUC)

	chatController := controllers.NewRAGChatController(chatUC, mqttSvc)
	chatController.StartMqttSubscription()

	// Setup Routes
	routes.SetupRAGRoutes(
		protected,
		controllers.NewRAGTranslateController(translateUC),
		controllers.NewRAGSummaryController(summaryUC),
		controllers.NewRAGStatusController(statusUC),
		chatController,
		controllers.NewRAGControlController(controlUC),
	)

	return refineUC
}
