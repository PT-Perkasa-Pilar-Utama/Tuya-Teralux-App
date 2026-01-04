package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid Bearer Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		var capturedToken string
		r.Use(AuthMiddleware())
		r.GET("/test", func(c *gin.Context) {
			token, exists := c.Get("access_token")
			if exists {
				capturedToken = token.(string)
			}
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token-123")
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if capturedToken != "test-token-123" {
			t.Errorf("Expected token 'test-token-123', got '%s'", capturedToken)
		}
	})

	t.Run("Token Without Bearer Prefix", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		var capturedToken string
		r.Use(AuthMiddleware())
		r.GET("/test", func(c *gin.Context) {
			token, exists := c.Get("access_token")
			if exists {
				capturedToken = token.(string)
			}
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "plain-token")
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if capturedToken != "plain-token" {
			t.Errorf("Expected token 'plain-token', got '%s'", capturedToken)
		}
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No Authorization header
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Invalid Authorization Header Format", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer token extra-part")
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("With Tuya UID Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		var capturedUID string
		r.Use(AuthMiddleware())
		r.GET("/test", func(c *gin.Context) {
			uid, exists := c.Get("tuya_uid")
			if exists {
				capturedUID = uid.(string)
			}
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-TUYA-UID", "user-123")
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if capturedUID != "user-123" {
			t.Errorf("Expected UID 'user-123', got '%s'", capturedUID)
		}
	})

	t.Run("Without Tuya UID Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		var uidExists bool
		r.Use(AuthMiddleware())
		r.GET("/test", func(c *gin.Context) {
			_, uidExists = c.Get("tuya_uid")
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		// No X-TUYA-UID header
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if uidExists {
			t.Error("Expected tuya_uid to not be set in context")
		}
	})
}
