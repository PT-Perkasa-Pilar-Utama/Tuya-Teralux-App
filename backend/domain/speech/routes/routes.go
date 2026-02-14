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
	whisperCppController *controllers.SpeechTranscribeWhisperCppController,
	ppuController *controllers.SpeechTranscribePPUController,
) {
	api := rg.Group("/api/speech")
	{
		// Transcription routes (unified fallback)
		api.POST("/transcribe", transcribeController.Transcribe)
		api.GET("/transcribe/:transcribe_id", statusController.GetStatus)

		// Whisper.cpp transcription route (async)
		api.POST("/transcribe/whisper/cpp", whisperCppController.TranscribeWhisperCpp)

		// Whisper proxy routes (external Outsystems integration) - kept for explicit access
		api.POST("/transcribe/ppu", ppuController.TranscribePPU)

	}
}
