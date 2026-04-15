package models

import (
	"path/filepath"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/providers"
	commonServices "sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineControllers "sensio/domain/models/pipeline/controllers"
	pipelinedtos "sensio/domain/models/pipeline/dtos"
	pipelineRoutes "sensio/domain/models/pipeline/routes"
	pipelineUsecases "sensio/domain/models/pipeline/usecases"
	ragControllers "sensio/domain/models/rag/controllers"
	ragdtos "sensio/domain/models/rag/dtos"
	ragRoutes "sensio/domain/models/rag/routes"
	ragServices "sensio/domain/models/rag/services"
	ragSkills "sensio/domain/models/rag/skills"
	ragOrchestrator "sensio/domain/models/rag/skills/orchestrator"
	ragUsecases "sensio/domain/models/rag/usecases"
	whisperControllers "sensio/domain/models/whisper/controllers"
	whisperDtos "sensio/domain/models/whisper/dtos"
	whisperRoutes "sensio/domain/models/whisper/routes"
	whisperUsecases "sensio/domain/models/whisper/usecases"
	recordingUsecases "sensio/domain/recordings/usecases"
	terminalRepositories "sensio/domain/terminal/terminal/repositories"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"time"

	"github.com/gin-gonic/gin"
)

// providerResolverTerminalRepoWrapper adapts ITerminalRepository to providers.TerminalRepository
type providerResolverTerminalRepoWrapper struct {
	repo terminalRepositories.ITerminalRepository
}

func (w *providerResolverTerminalRepoWrapper) GetByID(id string) (*providers.Terminal, error) {
	term, err := w.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &providers.Terminal{
		AiProvider:      term.AiProvider,
		AiEngineProfile: term.AiEngineProfile,
	}, nil
}

func (w *providerResolverTerminalRepoWrapper) GetByMacAddress(macAddress string) (*providers.Terminal, error) {
	term, err := w.repo.GetByMacAddress(macAddress)
	if err != nil {
		return nil, err
	}
	return &providers.Terminal{
		AiProvider:      term.AiProvider,
		AiEngineProfile: term.AiEngineProfile,
	}, nil
}

// InitModule initializes the consolidated models module (Whisper, RAG, and Pipeline).
func InitModule(
	protected *gin.RouterGroup,
	cfg *utils.Config,
	badger *infrastructure.BadgerService,
	vectorSvc *infrastructure.VectorService,
	tuyaAuth tuyaUsecases.TuyaAuthUseCase,
	tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor,
	mqttSvc *infrastructure.MqttService,
	terminalRepo terminalRepositories.ITerminalRepository,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
) (whisperUsecases.TranscribeUseCase, whisperUsecases.UploadSessionUseCase, ragUsecases.RefineUseCase, ragUsecases.TranslateUseCase, ragUsecases.SummaryUseCase) {

	// 1. Initialize RAG Sub-module
	// Initialize all provider services upfront for provider resolution
	geminiService, openaiService, groqService, orionService := providers.GetProviderServices(cfg)

	// Log provider direct upload limits at startup for observability
	utils.LogInfo("Startup: Provider direct upload limits | Gemini: %d MB | OpenAI: %d MB | Groq: %d MB | Orion: %d MB",
		commonServices.GeminiDirectUploadLimitBytes/1024/1024,
		commonServices.OpenAIDirectUploadLimitBytes/1024/1024,
		commonServices.GroqDirectUploadLimitBytes/1024/1024,
		commonServices.OrionDirectUploadLimitBytes/1024/1024,
	)

	// Create provider resolver for terminal-specific provider selection
	// Wrap terminalRepo to match the interface expected by ProviderResolver
	providerResolverRepo := &providerResolverTerminalRepoWrapper{terminalRepo}
	providerResolver := providers.NewProviderResolver(
		cfg,
		geminiService,
		openaiService,
		groqService,
		orionService,
		providerResolverRepo,
	)

	// Get default provider for backward compatibility
	defaultResolved := providerResolver.ResolveDefault()

	// CRITICAL: Fail fast at startup if no remote providers are configured
	// This prevents runtime nil pointer panics from misconfigured providers
	// Note: Local fallback is no longer used in default flow - only remote providers
	if defaultResolved.LLM == nil || defaultResolved.WhisperClient == nil {
		utils.LogError("Module Init: Provider resolution failed - no remote providers configured")
		utils.LogError("Module Init: Please set at least one of: OPENAI_API_KEY, GEMINI_API_KEY, GROQ_API_KEY, ORION_API_KEY")
		utils.LogError("Module Init: Or set LLM_PROVIDER to a valid provider (gemini/openai/groq/orion)")
		panic("Provider configuration error: no remote providers available")
	}

	// Get health-aware resolver for candidate-based selection
	healthAwareResolver := providerResolver.GetHealthAwareResolver()
	if healthAwareResolver != nil {
		candidates := healthAwareResolver.GetRemoteCandidates()
		utils.LogInfo("Startup: Health-aware resolver initialized | remote_candidates=%v | preferred_provider=%s", candidates, cfg.LLMProvider)
	} else {
		utils.LogWarn("Startup: Health-aware resolver not available")
	}

	ragLlmClient := defaultResolved.LLM

	ragStore := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()
	ragCache := tasks.NewBadgerTaskCacheFromService(badger, "cache:rag:task:")

	skillRegistry := ragSkills.NewSkillRegistry()
	basePath := "."
	if envPath := utils.FindEnvFile(); envPath != "" {
		basePath = filepath.Dir(envPath)
	}
	baseOrch := ragOrchestrator.NewBaseOrchestrator()
	controlOrch := ragOrchestrator.NewControlOrchestrator(tuyaExecutor, tuyaAuth)
	summaryOrch := ragOrchestrator.NewSummaryOrchestrator()

	skillsDir := filepath.Join(basePath, "domain", "models", "rag", "skills", "definitions")
	orchestratorResolver := func(name string) ragSkills.MarkdownOrchestrator {
		switch name {
		case "Control":
			return controlOrch
		case "Summary":
			return summaryOrch
		default:
			return baseOrch
		}
	}

	if err := ragSkills.LoadSkillsFromDirectory(skillsDir, skillRegistry, orchestratorResolver); err != nil {
		utils.LogError("RAG: Failed to load skills: %v", err)
	}

	summarySkill, _ := skillRegistry.Get("Summary")
	refineSkill, _ := skillRegistry.Get("Refine")
	translateSkill, _ := skillRegistry.Get("Translation")
	controlSkill, _ := skillRegistry.Get("Control")
	guardSkill, _ := skillRegistry.Get("Guard")
	chunkSkill, _ := skillRegistry.Get("ChunkSummary")
	structuredExtractionSkill, _ := skillRegistry.Get("StructuredExtraction")

	// Note: fallbackLLM is nil - default flow uses health-aware remote provider chain
	refineUC := ragUsecases.NewRefineUseCase(ragLlmClient, nil, cfg, refineSkill, providerResolver)
	translateUC := ragUsecases.NewTranslateUseCase(ragLlmClient, nil, cfg, ragCache, ragStore, mqttSvc, translateSkill, providerResolver)
	guardOrch := ragOrchestrator.NewGuardOrchestrator(guardSkill)
	fastIntentRouter := ragOrchestrator.NewFastIntentRouter()
	decisionEngine := ragOrchestrator.NewAssistantDecisionEngine(ragLlmClient)
	router := ragOrchestrator.NewRouter(skillRegistry, translateUC, guardOrch)
	pdfRenderer := ragServices.NewHTMLSummaryPDFRenderer()
	bigExternalService := commonServices.NewDeviceInfoExternalService()
	summaryUC := ragUsecases.NewSummaryUseCase(ragLlmClient, nil, cfg, ragCache, ragStore, pdfRenderer, bigExternalService, mqttSvc, summarySkill, chunkSkill, structuredExtractionSkill, providerResolver)
	ragStatusUC := tasks.NewGenericStatusUseCase(ragCache, ragStore)
	controlUC := ragUsecases.NewControlUseCase(ragLlmClient, nil, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth, controlSkill, providerResolver)
	chatUC := ragUsecases.NewChatUseCase(ragLlmClient, nil, cfg, badger, vectorSvc, guardOrch, fastIntentRouter, decisionEngine, providerResolver, controlUC, router)

	chatController := ragControllers.NewRAGChatController(chatUC, mqttSvc, terminalRepo)
	if err := chatController.StartMqttSubscription(); err != nil {
		utils.LogError("RAG module MQTT subscription failed: %v", err)
	}

	geminiRagRawUC := ragUsecases.NewQueryGeminiModelUseCase(geminiService)
	openaiRagRawUC := ragUsecases.NewQueryOpenAIModelUseCase(openaiService)
	groqRagRawUC := ragUsecases.NewQueryGroqModelUseCase(groqService)
	orionRagRawUC := ragUsecases.NewQueryOrionModelUseCase(orionService)

	ragRoutes.SetupRAGRoutes(
		protected,
		ragControllers.NewRAGTranslateController(translateUC),
		ragControllers.NewRAGSummaryController(summaryUC),
		ragControllers.NewRAGStatusController(ragStatusUC),
		chatController,
		ragControllers.NewRAGControlController(controlUC),
		ragControllers.NewRAGModelsGeminiController(geminiRagRawUC),
		ragControllers.NewRAGModelsOpenAIController(openaiRagRawUC),
		ragControllers.NewRAGModelsGroqController(groqRagRawUC),
		ragControllers.NewRAGModelsOrionController(orionRagRawUC),
	)

	// 2. Initialize Whisper Sub-module
	// Use provider resolver for Whisper client - no longer fixed to local
	// Default resolved provider includes both LLM and Whisper clients
	defaultWhisperClient := defaultResolved.WhisperClient

	whisperCache := tasks.NewBadgerTaskCacheFromService(badger, "cache:transcribe:task:")
	whisperStore := tasks.NewStatusStore[whisperDtos.AsyncTranscriptionStatusDTO]()

	// Initialize ASR Quality Gate components
	audioAnalyzer := utils.NewAudioAnalyzer()
	transcriptValidator := utils.NewTranscriptValidator()

	transcribeUC := whisperUsecases.NewTranscribeUseCase(defaultWhisperClient, refineUC, whisperStore, whisperCache, cfg, mqttSvc, providerResolver, audioAnalyzer, transcriptValidator)
	// Inject all provider services for health-aware fallback chain
	geminiWhisperModelUC := whisperUsecases.NewTranscribeGeminiModelUseCase(geminiService, whisperStore, whisperCache, cfg)
	openaiWhisperModelUC := whisperUsecases.NewTranscribeOpenAIModelUseCase(openaiService, whisperStore, whisperCache, cfg)
	groqWhisperModelUC := whisperUsecases.NewTranscribeGroqModelUseCase(groqService, whisperStore, whisperCache, cfg)
	orionWhisperModelUC := whisperUsecases.NewTranscribeOrionModelUseCase(orionService, whisperStore, whisperCache, cfg)
	uploadSessionUC := whisperUsecases.NewUploadSessionUseCase(badger, cfg)
	whisperStatusUC := tasks.NewGenericStatusUseCase(whisperCache, whisperStore)

	if cfg.EnableChunkUpload {
		cleanupInterval, err := time.ParseDuration(cfg.ChunkUploadCleanupInterval)
		if err != nil {
			cleanupInterval = 10 * time.Minute
		}
		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for {
				select {
				case now := <-ticker.C:
					count, err := uploadSessionUC.CleanupExpiredSessions(now)
					if err != nil {
						utils.LogError("Whisper: Upload session cleanup failed: %v", err)
					} else if count > 0 {
						utils.LogInfo("Whisper: Cleaned up %d expired upload sessions", count)
					}
				}
			}
		}()
	}

	transcribeController := whisperControllers.NewWhisperTranscribeController(transcribeUC, saveRecordingUC, uploadSessionUC, cfg, mqttSvc)
	if err := transcribeController.StartMqttSubscription(); err != nil {
		utils.LogError("Whisper module MQTT subscription failed: %v", err)
	}
	whisperStatusController := whisperControllers.NewWhisperTranscribeStatusController(whisperStatusUC)
	whisperUploadSessionController := whisperControllers.NewUploadSessionController(uploadSessionUC, transcribeUC)

	geminiWhisperController := whisperControllers.NewWhisperModelsGeminiController(geminiWhisperModelUC, saveRecordingUC, cfg)
	openaiWhisperController := whisperControllers.NewWhisperModelsOpenAIController(openaiWhisperModelUC, saveRecordingUC, cfg)
	groqWhisperController := whisperControllers.NewWhisperModelsGroqController(groqWhisperModelUC, saveRecordingUC, cfg)
	orionWhisperController := whisperControllers.NewWhisperModelsOrionController(orionWhisperModelUC, saveRecordingUC, cfg)

	whisperRoutes.SetupWhisperRoutes(
		protected,
		transcribeController,
		whisperStatusController,
		geminiWhisperController,
		openaiWhisperController,
		groqWhisperController,
		orionWhisperController,
		whisperUploadSessionController,
	)

	// 3. Initialize Pipeline Sub-module
	pipelineStore := tasks.NewStatusStore[pipelinedtos.PipelineStatusDTO]()
	pipelineCache := tasks.NewBadgerTaskCacheFromService(badger, "cache:pipeline:task:")

	pipelineUC := pipelineUsecases.NewPipelineUseCase(transcribeUC, translateUC, summaryUC, pipelineCache, pipelineStore, mqttSvc)
	pipelineStatusUC := tasks.NewGenericStatusUseCase(pipelineCache, pipelineStore)
	pipelineCtrl := pipelineControllers.NewPipelineController(pipelineUC, pipelineStatusUC, saveRecordingUC, uploadSessionUC, cfg)

	pipelineRoutes.SetupPipelineRoutes(protected, pipelineCtrl)

	return transcribeUC, uploadSessionUC, refineUC, translateUC, summaryUC
}
