package rag

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	commonServices "teralux_app/domain/common/services"
	"teralux_app/domain/rag/controllers"
	ragdtos "teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/routes"
	"teralux_app/domain/rag/services"
	"teralux_app/domain/rag/skills"
	"teralux_app/domain/rag/usecases"
	"teralux_app/domain/rag/utilities"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes RAG module with protected router group, configuration and optional persistence.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, tuyaAuth tuyaUsecases.TuyaAuthUseCase, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, mqttSvc *infrastructure.MqttService) usecases.RefineUseCase {
	// Initialize Dependencies (Services)
	orionService := commonServices.NewOrionService(cfg)
	geminiService := commonServices.NewGeminiService(cfg)


	// Select LLM Client based on configuration
	var llmClient utilities.LLMClient

	if cfg.LLMProvider == "gemini" {
		utils.LogInfo("RAG: Using Gemini as LLM Provider")
		llmClient = geminiService
	} else if cfg.LLMProvider == "orion" {
		utils.LogInfo("RAG: Using Orion as LLM Provider")
		llmClient = orionService
	} else {
		utils.LogFatal("RAG: Invalid or missing LLM_PROVIDER. Set it to 'gemini' or 'orion'. Detected: '%s'", cfg.LLMProvider)
		return nil // unreachable due to LogFatal likely os.Exit(1), but for safety
	}

	// Initialize Shared Store
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()
	cache := tasks.NewBadgerTaskCacheFromService(badger, "rag:task:")

	// Initialize Skills
	skillRegistry := skills.NewSkillRegistry()
	controlSkill := skills.NewControlSkill(tuyaExecutor, tuyaAuth)
	identitySkill := &skills.IdentitySkill{}
	translationSkill := &skills.TranslationSkill{}

	skillRegistry.Register(controlSkill)
	skillRegistry.Register(identitySkill)
	skillRegistry.Register(translationSkill)

	// Initialize Usecases
	refineUC := usecases.NewRefineUseCase(llmClient, cfg)
	translateUC := usecases.NewTranslateUseCase(llmClient, cfg, cache, store)

	orchestrator := skills.NewOrchestrator(skillRegistry, translateUC)
	pdfRenderer := services.NewMarotoSummaryPDFRenderer()
	summaryUC := usecases.NewSummaryUseCase(llmClient, cfg, cache, store, pdfRenderer)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	controlUC := usecases.NewControlUseCase(llmClient, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth)
	chatUC := usecases.NewChatUseCase(llmClient, cfg, badger, vectorSvc, orchestrator)

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
