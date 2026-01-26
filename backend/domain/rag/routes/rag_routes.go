package routes

import (
	"teralux_app/domain/rag/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRAGRoutes(r *gin.Engine, controller *controllers.RAGController) {
	v1 := r.Group("/v1")
	{
		v1.POST("/rag", controller.ProcessText)
		v1.GET("/rag/:task_id", controller.GetStatus)
	}
}
