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

	// Initialize Whisper Controller (Upload Session only)
	whisperUploadSessionCtrl := whisperControllers.NewUploadSessionController(whisperGrpcSvc)

	// Initialize Pipeline Controller
	pipelineCtrl := pipelineControllers.NewPipelineController(pipelineUC)

	// Setup Routes
	whisperRoutes.SetupWhisperRoutes(protected, whisperUploadSessionCtrl)

	pipelineRoutes.SetupPipelineRoutes(protected, pipelineCtrl)
}
