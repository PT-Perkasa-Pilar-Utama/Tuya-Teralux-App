package routes

import (
	"teralux_app/domain/rag/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRAGRoutes registers RAG endpoints under the protected router group.
func SetupRAGRoutes(rg *gin.RouterGroup, controller *controllers.RAGController) {
	api := rg.Group("/api/rag")
	{
		api.POST("/translate", controller.Translate)
		api.POST("/control", controller.Control)
		api.GET("/:task_id", controller.GetStatus)
	}
}
