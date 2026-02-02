package routes

import (
	"teralux_app/domain/speech/controllers"

	"github.com/gin-gonic/gin"
)

// SetupSpeechRoutes registers speech endpoints under the protected router group.
func SetupSpeechRoutes(rg *gin.RouterGroup, controller *controllers.TranscriptionController) {
	api := rg.Group("/api/speech")
	{
		api.POST("/transcribe", controller.HandleTranscribe)
		api.POST("/transcribe/long", controller.HandleTranscribeLong)
		api.POST("/mqtt/publish", controller.HandlePublishMqtt)
	}
}
