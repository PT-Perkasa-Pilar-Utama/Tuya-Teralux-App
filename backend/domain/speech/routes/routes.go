package routes

import (
	"teralux_app/domain/speech/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, controller *controllers.TranscriptionController) {
	api := router.Group("/")
	{
		api.POST("/transcribe", controller.HandleTranscribe)
	}
}
