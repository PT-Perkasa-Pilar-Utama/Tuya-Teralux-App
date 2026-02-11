package routes

import (
	"teralux_app/domain/speech/controllers"

	"github.com/gin-gonic/gin"
)

// SetupSpeechRoutes registers speech endpoints under the protected router group.
func SetupSpeechRoutes(rg *gin.RouterGroup, controller *controllers.TranscriptionController, whisperProxyController *controllers.WhisperProxyController) {
	api := rg.Group("/api/speech")
	{
		// Transcription routes (unified fallback)
		api.POST("/transcribe", controller.HandleProxyTranscribe)
		api.GET("/transcribe/:transcribe_id", controller.GetProxyTranscribeStatus)

		// Whisper.cpp transcription route (async)
		api.POST("/transcribe/whisper/cpp", controller.HandleWhisperCppTranscribe)

		// Whisper proxy routes (external Outsystems integration) - kept for explicit access
		api.POST("/transcribe/ppu", whisperProxyController.HandleProxyTranscribe)

	}
}
