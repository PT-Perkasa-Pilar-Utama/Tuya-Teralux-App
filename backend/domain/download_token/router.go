package download_token

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	downloadGroup := router.Group("/api/download")
	downloadGroup.POST("/token", handler.CreateToken)
	downloadGroup.GET("/:token", handler.ResolveToken)
}
