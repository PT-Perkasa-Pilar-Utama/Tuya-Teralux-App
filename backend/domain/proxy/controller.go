package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyController struct {
	targetBaseURL string
}

func NewProxyController() *ProxyController {
	// Default to local service if not provided in env
	target := os.Getenv("RAG_WHISPER_SERVICE_URL")
	if target == "" {
		target = "http://localhost:8000"
	}
	return &ProxyController{
		targetBaseURL: target,
	}
}

func (pc *ProxyController) ProxyRequest(c *gin.Context) {
	remote, err := url.Parse(pc.targetBaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid proxy target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host

		// Map /api/v1/rag/* -> /v1/rag/*
		// Map /api/v1/whisper/* -> /v1/speech/*
		// Map /api/models/rag/* -> /v1/rag/*
		// Map /api/models/whisper/* -> /v1/speech/*
		// Map /api/models/pipeline/* -> /v1/pipeline/*
		path := c.Request.URL.Path
		var newPath string
		if strings.HasPrefix(path, "/api/v1/rag") {
			newPath = strings.Replace(path, "/api/v1/rag", "/v1/rag", 1)
		} else if strings.HasPrefix(path, "/api/v1/whisper") {
			newPath = strings.Replace(path, "/api/v1/whisper", "/v1/speech", 1)
		} else if strings.HasPrefix(path, "/api/models/rag") {
			newPath = strings.Replace(path, "/api/models/rag", "/v1/rag", 1)
		} else if strings.HasPrefix(path, "/api/models/whisper") {
			newPath = strings.Replace(path, "/api/models/whisper", "/v1/speech", 1)
		} else if strings.HasPrefix(path, "/api/models/pipeline") {
			newPath = strings.Replace(path, "/api/models/pipeline", "/v1/pipeline", 1)
		} else {
			newPath = strings.TrimPrefix(path, "/api")
		}
		req.URL.Path = newPath
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (pc *ProxyController) RegisterRoutes(rg *gin.RouterGroup) {
	// New standard: /api/models/*
	models := rg.Group("/api/models")
	{
		rag := models.Group("/rag")
		{
			rag.Any("/*path", pc.ProxyRequest)
		}
		whisper := models.Group("/whisper")
		{
			whisper.Any("/*path", pc.ProxyRequest)
		}
		pipeline := models.Group("/pipeline")
		{
			pipeline.Any("/*path", pc.ProxyRequest)
		}
	}

	// Models info endpoint (Go native, not proxied)
	models.GET("", pc.GetModelsInfo)

	// Legacy support: /api/v1/* (backward compatibility)
	v1 := rg.Group("/api/v1")
	{
		rag := v1.Group("/rag")
		{
			rag.Any("/*path", pc.ProxyRequest)
		}
		whisper := v1.Group("/whisper")
		{
			whisper.Any("/*path", pc.ProxyRequest)
		}
	}

	// Legacy support: /api/rag and /api/whisper (backward compatibility)
	// This allows gradual migration from old paths to new /api/v1/* and /api/models/* paths
	legacy := rg.Group("/api")
	{
		rag := legacy.Group("/rag")
		{
			rag.Any("/*path", pc.ProxyRequest)
		}
		whisper := legacy.Group("/whisper")
		{
			whisper.Any("/*path", pc.ProxyRequest)
		}
	}
}

func (pc *ProxyController) GetModelsInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"models": []string{"rag", "whisper", "pipeline"},
		"status": "ok",
	})
}
