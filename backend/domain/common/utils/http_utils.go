package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetBaseURL extracts the base URL from the gin context.
// It handles standard Host and respects X-Forwarded-Proto/Host headers if behind a proxy.
func GetBaseURL(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil || ctx.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := ctx.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = ctx.Request.Host
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}
