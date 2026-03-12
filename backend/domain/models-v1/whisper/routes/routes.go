package routes

import (
	"sensio/domain/models-v1/whisper/controllers"
	"sensio/domain/models-v1/whisper/services"

	"github.com/gin-gonic/gin"
)

// SetupWhisperRoutes registers whisper endpoints under the protected router group.
func SetupWhisperRoutes(
	rg *gin.RouterGroup,
	uploadSessionController *controllers.UploadSessionController,
	whisperController *controllers.WhisperController,
) {
	// V1 API: /api/models/v1/whisper/*
	models := rg.Group("/api/models/v1/whisper")
	{
		models.POST("/transcribe", whisperController.Transcribe)
		models.GET("/transcribe/:transcribe_id", whisperController.GetStatus)

		// Upload session routes
		uploads := models.Group("/uploads")
		{
			uploads.POST("/sessions", uploadSessionController.CreateSession)
			uploads.PUT("/sessions/:id/chunks/:index", uploadSessionController.UploadChunk)
			uploads.GET("/sessions/:id", uploadSessionController.GetSessionStatus)
		}
	}
}

// SetupLegacyWhisperRoutes is a wrapper for SetupWhisperRoutes to match the expected signature in module.go if needed,
// but we should ideally update module.go to use the new SetupWhisperRoutes.
// For now, let's just make it compatible.
func SetupLegacyWhisperRoutes(
	rg *gin.RouterGroup,
	uploadSessionController *controllers.UploadSessionController,
	grpcSvc *services.GrpcWhisperService,
) {
	whisperCtrl := controllers.NewWhisperController(grpcSvc)
	SetupWhisperRoutes(rg, uploadSessionController, whisperCtrl)
}
