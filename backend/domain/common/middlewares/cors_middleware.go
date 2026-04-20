package middlewares

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"sensio/domain/common/utils"
)

// CorsMiddleware returns a Gin middleware that handles CORS.
func CorsMiddleware() gin.HandlerFunc {
	cfg := utils.GetConfig()
	allowedOrigins := strings.Split(cfg.AllowedOrigins, ",")
	if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && strings.TrimSpace(allowedOrigins[0]) == "") {
		allowedOrigins = []string{"http://localhost:3000", "http://localhost:8081"}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-KEY", "X-TUYA-UID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
