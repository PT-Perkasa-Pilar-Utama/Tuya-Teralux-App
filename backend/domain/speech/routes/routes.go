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


		// Transcription long routes (async) - kept for explicit access if needed
		api.POST("/transcribe/long", controller.HandleProxyTranscribeLong)

		// Whisper proxy routes (external Outsystems integration) - kept for explicit access
		api.POST("/transcribe/ppu", whisperProxyController.HandleProxyTranscribe)

		// Other routes
		api.POST("/mqtt/publish", controller.HandlePublishMqtt)
	}
}