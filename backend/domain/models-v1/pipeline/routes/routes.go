package routes

import (
	"sensio/domain/models-v1/pipeline/controllers"

	"github.com/gin-gonic/gin"
)

// SetupPipelineRoutes registers pipeline endpoints under the protected router group.
func SetupPipelineRoutes(
	protected *gin.RouterGroup,
	pipelineCtrl *controllers.PipelineController,
) {
	// V1 API: /api/v1/models/pipeline/*
	models := protected.Group("/api/v1/models/pipeline")
	{
		models.POST("/job", pipelineCtrl.ExecuteJob)
		models.POST("/job/by-upload", pipelineCtrl.ExecuteJobByUpload)
		models.GET("/status/:task_id", pipelineCtrl.GetStatus)
	}
}
