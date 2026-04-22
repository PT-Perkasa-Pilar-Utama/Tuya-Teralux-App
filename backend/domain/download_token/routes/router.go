package routes

import (
	"github.com/gin-gonic/gin"

	"sensio/domain/download_token/controllers"
)

func RegisterRoutes(router *gin.RouterGroup, handler *controllers.Handler) {
	downloadGroup := router.Group("/api/download")
	downloadGroup.POST("/token", handler.CreateToken)
	downloadGroup.GET("/resolve", handler.ResolveToken)
}
