package routes

import (
	"teralux_app/domain/speech/controllers"

	"github.com/gin-gonic/gin"
)

// SetupSpeechRoutes registers speech endpoints under the protected router group.
func SetupSpeechRoutes(
	rg *gin.RouterGroup,
	transcribeController *controllers.SpeechTranscribeController,
	statusController *controllers.SpeechTranscribeStatusController,
	geminiController *controllers.SpeechModelsGeminiController,
	openaiController *controllers.SpeechModelsOpenAIController,
	groqController *controllers.SpeechModelsGroqController,
	orionController *controllers.SpeechModelsOrionController,
	cppModelController *controllers.SpeechModelsWhisperCppController,
) {
	api := rg.Group("/api/speech")
	{
		// Transcription routes (unified fallback)
		api.POST("/transcribe", transcribeController.Transcribe)
		api.GET("/transcribe/:transcribe_id", statusController.GetStatus)

		// Model-specific transcription routes (direct, async, no default refinement)
		models := api.Group("/models")
		{
			models.POST("/gemini", geminiController.Transcribe)
			models.POST("/openai", openaiController.Transcribe)
			models.POST("/groq", groqController.Transcribe)
			models.POST("/orion", orionController.Transcribe)
			models.POST("/whisper/cpp", cppModelController.Transcribe)
		}

	}
}
