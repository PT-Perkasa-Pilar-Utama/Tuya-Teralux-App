package routes

import (
	"sensio/domain/models/pipeline/controllers"

	"github.com/gin-gonic/gin"
)

func SetupPipelineRoutes(
	protected *gin.RouterGroup,
	pipelineCtrl *controllers.PipelineController,
) {
	// New standard: /api/models/pipeline/*
	models := protected.Group("/api/models/pipeline")
	{
		models.POST("/job", pipelineCtrl.ExecuteJob)
		models.POST("/job/by-upload", pipelineCtrl.ExecuteJobByUpload)
		models.GET("/status/:task_id", pipelineCtrl.GetStatus)
		models.DELETE("/status/:task_id", pipelineCtrl.CancelTask)
	}

	// Legacy support: /api/pipeline/* (backward compatibility)
	legacy := protected.Group("/api/pipeline")
	{
		legacy.POST("/job", pipelineCtrl.ExecuteJob)
		legacy.POST("/job/by-upload", pipelineCtrl.ExecuteJobByUpload)
		legacy.GET("/status/:task_id", pipelineCtrl.GetStatus)
		legacy.DELETE("/status/:task_id", pipelineCtrl.CancelTask)
	}
}
