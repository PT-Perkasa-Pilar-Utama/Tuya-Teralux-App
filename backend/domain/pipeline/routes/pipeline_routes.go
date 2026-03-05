package routes

import (
	"sensio/domain/pipeline/controllers"

	"github.com/gin-gonic/gin"
)

func SetupPipelineRoutes(
	protected *gin.RouterGroup,
	pipelineCtrl *controllers.PipelineController,
) {
	pipeline := protected.Group("/api/pipeline")
	{
		pipeline.POST("/job", pipelineCtrl.ExecuteJob)
		pipeline.POST("/job/by-upload", pipelineCtrl.ExecuteJobByUpload)
		pipeline.GET("/status/:task_id", pipelineCtrl.GetStatus)
	}
}
