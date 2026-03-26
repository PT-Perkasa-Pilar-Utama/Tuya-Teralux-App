package routes

import (
	"sensio/domain/models/rag/controllers"

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
) {
	// New standard: /api/models/rag/*
	models := rg.Group("/api/models/rag")
	{
		models.POST("/translate", translateController.Translate)
		models.POST("/summary", summaryController.Summary)
		models.POST("/chat", chatController.Chat)
		models.POST("/control", controlController.Control)
		models.GET("/:task_id", statusController.GetStatus)

		// Model-specific RAG routes (LLM providers)
		models.POST("/gemini", geminiModelCtrl.Query)
		models.POST("/openai", openaiModelCtrl.Query)
		models.POST("/groq", groqModelCtrl.Query)
		models.POST("/orion", orionModelCtrl.Query)
	}

	// Legacy support: /api/rag/* (backward compatibility)
}
