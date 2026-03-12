package routes

import (
	"sensio/domain/models-v1/whisper/controllers"

	"github.com/gin-gonic/gin"
)

// SetupWhisperRoutes registers whisper upload session endpoints under the protected router group.
func SetupWhisperRoutes(
	rg *gin.RouterGroup,
	uploadSessionController *controllers.UploadSessionController,
) {
	// V1 API: /api/v1/models/whisper/*
	models := rg.Group("/api/v1/models/whisper")
	{
		// Upload session routes
		uploads := models.Group("/uploads")
		{
			uploads.POST("/sessions", uploadSessionController.CreateSession)
			uploads.PUT("/sessions/:id/chunks/:index", uploadSessionController.UploadChunk)
			uploads.GET("/sessions/:id", uploadSessionController.GetSessionStatus)
		}
	}
}
