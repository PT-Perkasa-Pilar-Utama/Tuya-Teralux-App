package routes

import (
	"sensio/domain/models-v1/rag/controllers"

	"github.com/gin-gonic/gin"
)

// SetupLegacyRAGRoutes registers legacy RAG endpoints under the protected router group.
func SetupLegacyRAGRoutes(
	rg *gin.RouterGroup,
	ctrl *controllers.RAGController,
) {
	// V1 API: /api/models/v1/rag/*
	models := rg.Group("/api/models/v1/rag")
	{
		models.POST("/translate", ctrl.Translate)
		models.POST("/summary", ctrl.Summary)
		models.POST("/chat", ctrl.Chat)
		models.POST("/control", ctrl.Control)
		models.GET("/:task_id", ctrl.GetStatus)
	}
}
