package routes

import (
	"sensio/domain/models-v1/pipeline/controllers"

	"github.com/gin-gonic/gin"
)

// SetupPipelineRoutes registers pipeline endpoints under the protected router group.
func SetupPipelineRoutes(rg *gin.RouterGroup, ctrl *controllers.PipelineController) {
	// V1 API: /api/models/v1/pipeline/*
	models := rg.Group("/api/models/v1/pipeline")
	{
		models.POST("/job", ctrl.ExecuteJob)
		models.GET("/status/:id", ctrl.GetStatus)
	}
}
