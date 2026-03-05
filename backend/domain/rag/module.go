package rag

import (
	"path/filepath"
	"sensio/domain/common/infrastructure"
	commonServices "sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/rag/controllers"
	ragdtos "sensio/domain/rag/dtos"
	"sensio/domain/rag/routes"
	"sensio/domain/rag/services"
	"sensio/domain/rag/skills"
	"sensio/domain/rag/usecases"
	tuyaUsecases "sensio/domain/tuya/usecases"

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
	cache := tasks.NewBadgerTaskCacheFromService(badger, "cache:rag:task:")

	// Initialize Skills from Markdown definitions
	skillRegistry := skills.NewSkillRegistry()
	basePath := "."
	if envPath := utils.FindEnvFile(); envPath != "" {
		basePath = filepath.Dir(envPath)
	}
	skillsDir := filepath.Join(basePath, "domain", "rag", "skills", "definitions")
	if err := skills.LoadSkillsFromDirectory(skillsDir, skillRegistry, tuyaExecutor, tuyaAuth); err != nil {
		utils.LogError("RAG: Failed to load skills: %v", err)
	}

	// Retrieve specific skills for usecases (safe default to nil if not found)
	summarySkill, _ := skillRegistry.Get("Summary")
	refineSkill, _ := skillRegistry.Get("Refine")
	translateSkill, _ := skillRegistry.Get("Translation")
	controlSkill, _ := skillRegistry.Get("Control")

	// Initialize Usecases
	refineUC := usecases.NewRefineUseCase(llmClient, llamaService, cfg, refineSkill)
	translateUC := usecases.NewTranslateUseCase(llmClient, llamaService, cfg, cache, store, mqttSvc, translateSkill)

	orchestrator := skills.NewOrchestrator(skillRegistry, translateUC)
	pdfRenderer := services.NewHTMLSummaryPDFRenderer()
	bigExternalService := commonServices.NewBigExternalService()
	summaryUC := usecases.NewSummaryUseCase(llmClient, llamaService, cfg, cache, store, pdfRenderer, bigExternalService, mqttSvc, summarySkill)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	controlUC := usecases.NewControlUseCase(llmClient, llamaService, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth, controlSkill)
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
