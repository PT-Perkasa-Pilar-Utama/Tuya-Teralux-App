package models_v1

import (
	"sensio/domain/common/utils"
	pipelineControllers "sensio/domain/models-v1/pipeline/controllers"
	pipelineRoutes "sensio/domain/models-v1/pipeline/routes"
	pipelineServices "sensio/domain/models-v1/pipeline/services"
	pipelineUsecases "sensio/domain/models-v1/pipeline/usecases"
	whisperControllers "sensio/domain/models-v1/whisper/controllers"
	whisperRoutes "sensio/domain/models-v1/whisper/routes"
	whisperServices "sensio/domain/models-v1/whisper/services"
	ragControllers "sensio/domain/models-v1/rag/controllers"
	ragRoutes "sensio/domain/models-v1/rag/routes"
	ragServices "sensio/domain/models-v1/rag/services"

	"github.com/gin-gonic/gin"
)

// InitModule initializes the models-v1 module with the protected router group.
// This module acts as a proxy to Python services via gRPC.
// Only Pipeline and Whisper Upload Session routes are kept.
func InitModule(protected *gin.RouterGroup, cfg *utils.Config) {
	// Initialize Whisper gRPC Service
	whisperGrpcSvc, err := whisperServices.NewGrpcWhisperService(cfg)
	if err != nil {
		utils.LogError("Failed to initialize Whisper gRPC service: %v", err)
		return
	}

	// Initialize Pipeline Service (HTTP to Python)
	pipelinePythonSvc := pipelineServices.NewPythonPipelineService(cfg)

	// Initialize Pipeline Usecases
	pipelineUC := pipelineUsecases.NewPipelineUseCase(pipelinePythonSvc)

	// Initialize Whisper Controllers
	whisperUploadSessionCtrl := whisperControllers.NewUploadSessionController(whisperGrpcSvc)
	whisperCtrl := whisperControllers.NewWhisperController(whisperGrpcSvc)

	// Initialize Pipeline Controller
	pipelineCtrl := pipelineControllers.NewPipelineController(pipelineUC)

	// Initialize RAG components (Legacy V1 via REST)
	ragSvc := ragServices.NewPythonRAGService(cfg)
	ragCtrl := ragControllers.NewRAGController(ragSvc)

	// Setup Routes
	// All routes now use prefix /api/models/v1/domain
	whisperRoutes.SetupWhisperRoutes(protected, whisperUploadSessionCtrl, whisperCtrl)

	pipelineRoutes.SetupPipelineRoutes(protected, pipelineCtrl)

	// Setup Legacy V1 Routes (Direct Python service access)
	// Routes: /api/models/v1/rag/*
	ragRoutes.SetupLegacyRAGRoutes(protected, ragCtrl)
}
