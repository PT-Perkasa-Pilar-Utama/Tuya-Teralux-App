package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApiKeyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid API Key", func(t *testing.T) {
		// Set valid API key in environment
		_ = os.Setenv("API_KEY", "test-api-key-123")
		defer func() { _ = os.Unsetenv("API_KEY") }()

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(ApiKeyMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-KEY", "test-api-key-123")
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		_ = os.Setenv("API_KEY", "correct-key")
		defer func() { _ = os.Unsetenv("API_KEY") }()

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(ApiKeyMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-KEY", "wrong-key")
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Missing API Key Header", func(t *testing.T) {
		_ = os.Setenv("API_KEY", "test-key")
		defer func() { _ = os.Unsetenv("API_KEY") }()

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(ApiKeyMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No X-API-KEY header set
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Server API Key Not Configured", func(t *testing.T) {
		// Ensure API_KEY is not set
		_ = os.Unsetenv("API_KEY")

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(ApiKeyMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-KEY", "any-key")
		c.Request = req

		r.ServeHTTP(w, req)

		// When API_KEY is empty, it's treated as invalid key (401)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}
