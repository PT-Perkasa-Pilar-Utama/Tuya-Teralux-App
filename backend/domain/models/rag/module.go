package rag

import (
	"path/filepath"
	"sensio/domain/common/infrastructure"
	commonServices "sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/controllers"
	ragdtos "sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/routes"
	"sensio/domain/models/rag/services"
	"sensio/domain/models/rag/skills"
	"sensio/domain/models/rag/skills/orchestrator"
	"sensio/domain/models/rag/usecases"
	terminal_repositories "sensio/domain/terminal/repositories"
	tuyaUsecases "sensio/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// InitModule initializes RAG module with protected router group, configuration and optional persistence.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config, badger *infrastructure.BadgerService, vectorSvc *infrastructure.VectorService, tuyaAuth tuyaUsecases.TuyaAuthUseCase, tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, mqttSvc *infrastructure.MqttService, terminalRepo terminal_repositories.ITerminalRepository) (usecases.RefineUseCase, usecases.TranslateUseCase, usecases.SummaryUseCase) {
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
		return nil, nil, nil // unreachable due to LogFatal likely os.Exit(1), but for safety
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
	// Initialize Specialized Orchestrators
	baseOrch := orchestrator.NewBaseOrchestrator()
	controlOrch := orchestrator.NewControlOrchestrator(tuyaExecutor, tuyaAuth)
	summaryOrch := orchestrator.NewSummaryOrchestrator()

	skillsDir := filepath.Join(basePath, "domain", "rag", "skills", "definitions")
	orchestratorResolver := func(name string) skills.MarkdownOrchestrator {
		switch name {
		case "Control":
			return controlOrch
		case "Summary":
			return summaryOrch
		default:
			return baseOrch
		}
	}

	if err := skills.LoadSkillsFromDirectory(skillsDir, skillRegistry, orchestratorResolver); err != nil {
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

	// Initialize Guard Orchestrator
	guardSkill, _ := skillRegistry.Get("Guard")
	guardOrch := orchestrator.NewGuardOrchestrator(guardSkill)
	chunkSkill, _ := skillRegistry.Get("ChunkSummary")

	router := orchestrator.NewRouter(skillRegistry, translateUC, guardOrch)
	pdfRenderer := services.NewHTMLSummaryPDFRenderer()
	bigExternalService := commonServices.NewBigExternalService()
	summaryUC := usecases.NewSummaryUseCase(llmClient, llamaService, cfg, cache, store, pdfRenderer, bigExternalService, mqttSvc, summarySkill, chunkSkill)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	controlUC := usecases.NewControlUseCase(llmClient, llamaService, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth, controlSkill)
	chatUC := usecases.NewChatUseCase(llmClient, llamaService, cfg, badger, vectorSvc, router)

	chatController := controllers.NewRAGChatController(chatUC, mqttSvc, terminalRepo)
	if err := chatController.StartMqttSubscription(); err != nil {
		utils.LogError("RAG module MQTT subscription failed: %v", err)
	}

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

	return refineUC, translateUC, summaryUC
}
