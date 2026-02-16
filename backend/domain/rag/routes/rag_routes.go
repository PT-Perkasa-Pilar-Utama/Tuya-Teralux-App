package routes

import (
	"teralux_app/domain/rag/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRAGRoutes registers RAG endpoints under the protected router group.
func SetupRAGRoutes(
	rg *gin.RouterGroup,
	translateController *controllers.RAGTranslateController,
	summaryController *controllers.RAGSummaryController,
	statusController *controllers.RAGStatusController,
	chatController *controllers.RAGChatController,
	controlController *controllers.RAGControlController,
) {
	api := rg.Group("/api/rag")
	{
		api.POST("/translate", translateController.Translate)
		api.POST("/summary", summaryController.Summary)
		api.POST("/chat", chatController.Chat)
		api.POST("/control", controlController.Control)
		api.GET("/:task_id", statusController.GetStatus)
	}
}
