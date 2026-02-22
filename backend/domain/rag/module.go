package rag

import (
	"teralux_app/domain/common/infrastructure"
	commonServices "teralux_app/domain/common/services"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/controllers"
	ragdtos "teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/routes"
	"teralux_app/domain/rag/services"
	"teralux_app/domain/rag/skills"
	"teralux_app/domain/rag/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes RAG module with protected router group, configuration and optional persistence.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, tuyaAuth tuyaUsecases.TuyaAuthUseCase, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, mqttSvc *infrastructure.MqttService) usecases.RefineUseCase {
	// Initialize Dependencies	// Services
	geminiService := commonServices.NewGeminiService(cfg)
	orionService := commonServices.NewOrionService(cfg)
	openaiService := commonServices.NewOpenAIService(cfg)
	groqService := commonServices.NewGroqService(cfg)
	llamaService := commonServices.NewLlamaLocalService(cfg)

	// Select LLM Client based on configuration
	var llmClient skills.LLMClient

	switch cfg.LLMProvider {
	case "gemini":
		utils.LogInfo("RAG: Using Gemini as LLM Provider")
		llmClient = geminiService
	case "orion":
		utils.LogInfo("RAG: Using Orion as LLM Provider")
		llmClient = orionService
	case "openai":
		utils.LogInfo("RAG: Using OpenAI as LLM Provider")
		llmClient = openaiService
	case "groq":
		utils.LogInfo("RAG: Using Groq as LLM Provider")
		llmClient = groqService
	case "local":
		utils.LogInfo("RAG: Using Local Llama (llama.cpp) as LLM Provider")
		llmClient = llamaService
	default:
		utils.LogFatal("RAG: Invalid or missing LLM_PROVIDER. Set it to 'gemini', 'orion', 'openai', 'groq', or 'local'. Detected: '%s'", cfg.LLMProvider)
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
	refineUC := usecases.NewRefineUseCase(llmClient, llamaService, cfg)
	translateUC := usecases.NewTranslateUseCase(llmClient, llamaService, cfg, cache, store)

	orchestrator := skills.NewOrchestrator(skillRegistry, translateUC)
	pdfRenderer := services.NewMarotoSummaryPDFRenderer()
	summaryUC := usecases.NewSummaryUseCase(llmClient, llamaService, cfg, cache, store, pdfRenderer)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	controlUC := usecases.NewControlUseCase(llmClient, llamaService, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth)
	chatUC := usecases.NewChatUseCase(llmClient, llamaService, cfg, badger, vectorSvc, orchestrator)

	chatController := controllers.NewRAGChatController(chatUC, mqttSvc)
	chatController.StartMqttSubscription()

	// Setup Usecases for Raw Models
	geminiRawUC := usecases.NewQueryGeminiModelUseCase(geminiService)
	openaiRawUC := usecases.NewQueryOpenAIModelUseCase(openaiService)
	groqRawUC := usecases.NewQueryGroqModelUseCase(groqService)
	orionRawUC := usecases.NewQueryOrionModelUseCase(orionService)
	llamaRawUC := usecases.NewQueryLlamaCppModelUseCase(llamaService)

	// Setup Routes
	routes.SetupRAGRoutes(
		protected,
		controllers.NewRAGTranslateController(translateUC),
		controllers.NewRAGSummaryController(summaryUC),
		controllers.NewRAGStatusController(statusUC),
		chatController,
		controllers.NewRAGControlController(controlUC),
		controllers.NewRAGModelsGeminiController(geminiRawUC),
		controllers.NewRAGModelsOpenAIController(openaiRawUC),
		controllers.NewRAGModelsGroqController(groqRawUC),
		controllers.NewRAGModelsOrionController(orionRawUC),
		controllers.NewRAGModelsLlamaCppController(llamaRawUC),
	)

	return refineUC
}
