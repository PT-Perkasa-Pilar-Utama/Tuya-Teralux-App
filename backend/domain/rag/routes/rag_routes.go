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
	geminiModelCtrl controllers.RAGModelsGeminiController,
	openaiModelCtrl controllers.RAGModelsOpenAIController,
	groqModelCtrl controllers.RAGModelsGroqController,
	orionModelCtrl controllers.RAGModelsOrionController,
	llamaCppModelCtrl controllers.RAGModelsLlamaCppController,
) {
	api := rg.Group("/api/rag")
	{
		api.POST("/translate", translateController.Translate)
		api.POST("/summary", summaryController.Summary)
		api.POST("/chat", chatController.Chat)
		api.POST("/control", controlController.Control)
		api.GET("/:task_id", statusController.GetStatus)
		
		models := rg.Group("/api/models")
		{
			models.POST("/gemini", geminiModelCtrl.Query)
			models.POST("/openai", openaiModelCtrl.Query)
			models.POST("/groq", groqModelCtrl.Query)
			models.POST("/orion", orionModelCtrl.Query)
			models.POST("/llama/cpp", llamaCppModelCtrl.Query)
		}
	}
}
