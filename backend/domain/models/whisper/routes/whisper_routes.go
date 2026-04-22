package routes

import (
	"sensio/domain/models/whisper/controllers"

	"github.com/gin-gonic/gin"
)

// SetupWhisperRoutes registers whisper endpoints under the protected router group.
func SetupWhisperRoutes(
	rg *gin.RouterGroup,
	transcribeController *controllers.WhisperTranscribeController,
	statusController *controllers.WhisperTranscribeStatusController,
	geminiController *controllers.WhisperModelsGeminiController,
	openaiController *controllers.WhisperModelsOpenAIController,
	groqController *controllers.WhisperModelsGroqController,
	orionController *controllers.WhisperModelsOrionController,
	uploadSessionController *controllers.UploadSessionController,
) {
	// New standard: /api/models/whisper/*
	models := rg.Group("/api/models/whisper")
	{
		// Transcription routes
		models.POST("/transcribe", transcribeController.Transcribe)
		models.POST("/transcribe/by-upload", transcribeController.TranscribeByUpload)
		models.GET("/transcribe/:transcribe_id", statusController.GetStatus)

		// Upload session routes
		uploads := models.Group("/uploads")
		{
			uploads.POST("/sessions", uploadSessionController.CreateSession)
			uploads.PUT("/sessions/:id/chunks/:index", uploadSessionController.UploadChunk)
			uploads.GET("/sessions/:id", uploadSessionController.GetSessionStatus)
		}

		// Model-specific transcription routes (LLM providers)
		models.POST("/gemini", geminiController.Transcribe)
		models.POST("/openai", openaiController.Transcribe)
		models.POST("/groq", groqController.Transcribe)
		models.POST("/orion", orionController.Transcribe)
	}

}
