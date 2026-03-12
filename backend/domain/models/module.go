package models

import (
	"path/filepath"
	"sensio/domain/common/infrastructure"
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
	geminiService := commonServices.NewGeminiService(cfg)
	orionService := commonServices.NewOrionService(cfg)
	openaiService := commonServices.NewOpenAIService(cfg)
	groqService := commonServices.NewGroqService(cfg)
	llamaService := commonServices.NewLlamaLocalService(cfg)

	var ragLlmClient ragSkills.LLMClient
	switch cfg.LLMProvider {
	case "gemini":
		utils.LogInfo("RAG: Using Gemini as LLM Provider")
		ragLlmClient = geminiService
	case "orion":
		utils.LogInfo("RAG: Using Orion as LLM Provider")
		ragLlmClient = orionService
	case "openai":
		utils.LogInfo("RAG: Using OpenAI as LLM Provider")
		ragLlmClient = openaiService
	case "groq":
		utils.LogInfo("RAG: Using Groq as LLM Provider")
		ragLlmClient = groqService
	case "local":
		utils.LogInfo("RAG: Using Local Llama (llama.cpp) as LLM Provider")
		ragLlmClient = llamaService
	default:
		utils.LogFatal("RAG: Invalid or missing LLM_PROVIDER for RAG.")
	}

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

	refineUC := ragUsecases.NewRefineUseCase(ragLlmClient, llamaService, cfg, refineSkill)
	translateUC := ragUsecases.NewTranslateUseCase(ragLlmClient, llamaService, cfg, ragCache, ragStore, mqttSvc, translateSkill)
	guardOrch := ragOrchestrator.NewGuardOrchestrator(guardSkill)
	router := ragOrchestrator.NewRouter(skillRegistry, translateUC, guardOrch)
	pdfRenderer := ragServices.NewHTMLSummaryPDFRenderer()
	bigExternalService := commonServices.NewBigExternalService()
	summaryUC := ragUsecases.NewSummaryUseCase(ragLlmClient, llamaService, cfg, ragCache, ragStore, pdfRenderer, bigExternalService, mqttSvc, summarySkill, chunkSkill)
	ragStatusUC := tasks.NewGenericStatusUseCase(ragCache, ragStore)
	controlUC := ragUsecases.NewControlUseCase(ragLlmClient, llamaService, cfg, vectorSvc, badger, tuyaExecutor, tuyaAuth, controlSkill)
	chatUC := ragUsecases.NewChatUseCase(ragLlmClient, llamaService, cfg, badger, vectorSvc, router)

	chatController := ragControllers.NewRAGChatController(chatUC, mqttSvc, terminalRepo)
	if err := chatController.StartMqttSubscription(); err != nil {
		utils.LogError("RAG module MQTT subscription failed: %v", err)
	}

	geminiRagRawUC := ragUsecases.NewQueryGeminiModelUseCase(geminiService)
	openaiRagRawUC := ragUsecases.NewQueryOpenAIModelUseCase(openaiService)
	groqRagRawUC := ragUsecases.NewQueryGroqModelUseCase(groqService)
	orionRagRawUC := ragUsecases.NewQueryOrionModelUseCase(orionService)
	llamaRagRawUC := ragUsecases.NewQueryLlamaCppModelUseCase(llamaService)

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
		ragControllers.NewRAGModelsLlamaCppController(llamaRagRawUC),
	)

	// 2. Initialize Whisper Sub-module
	localWhisperService := commonServices.NewWhisperLocalService(cfg)
	var whisperClient whisperUsecases.WhisperClient
	switch cfg.LLMProvider {
	case "gemini":
		utils.LogInfo("Whisper: Using Gemini Whisper (Multimodal)")
		whisperClient = geminiService
	case "openai":
		utils.LogInfo("Whisper: Using OpenAI Whisper")
		whisperClient = openaiService
	case "groq":
		utils.LogInfo("Whisper: Using Groq Whisper")
		whisperClient = groqService
	case "orion":
		if cfg.OrionWhisperBaseURL != "" {
			utils.LogInfo("Whisper: Using Remote Whisper (Orion)")
			whisperClient = orionService
		} else {
			utils.LogFatal("Whisper: LLM_PROVIDER is 'orion' but ORION_WHISPER_BASE_URL is not set.")
		}
	default:
		utils.LogFatal("Whisper: Invalid or missing LLM_PROVIDER for Whisper.")
	}

	whisperCache := tasks.NewBadgerTaskCacheFromService(badger, "cache:transcribe:task:")
	whisperStore := tasks.NewStatusStore[whisperDtos.AsyncTranscriptionStatusDTO]()

	transcribeUC := whisperUsecases.NewTranscribeUseCase(whisperClient, localWhisperService, refineUC, whisperStore, whisperCache, cfg, mqttSvc)
	geminiWhisperModelUC := whisperUsecases.NewTranscribeGeminiModelUseCase(geminiService, whisperStore, whisperCache, cfg)
	openaiWhisperModelUC := whisperUsecases.NewTranscribeOpenAIModelUseCase(openaiService, whisperStore, whisperCache, cfg)
	groqWhisperModelUC := whisperUsecases.NewTranscribeGroqModelUseCase(groqService, whisperStore, whisperCache, cfg)
	orionWhisperModelUC := whisperUsecases.NewTranscribeOrionModelUseCase(orionService, whisperStore, whisperCache, cfg)
	cppWhisperModelUC := whisperUsecases.NewTranscribeWhisperCppModelUseCase(localWhisperService, whisperStore, whisperCache, cfg)
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
	cppWhisperController := whisperControllers.NewWhisperModelsWhisperCppController(cppWhisperModelUC, saveRecordingUC, cfg)

	whisperRoutes.SetupWhisperRoutes(
		protected,
		transcribeController,
		whisperStatusController,
		geminiWhisperController,
		openaiWhisperController,
		groqWhisperController,
		orionWhisperController,
		cppWhisperController,
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
