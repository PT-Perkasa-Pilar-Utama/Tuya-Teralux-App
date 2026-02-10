package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"teralux_app/domain/common/utils"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockTokenProvider for testing
type MockTokenProvider struct{}

func (m *MockTokenProvider) GetTuyaAccessToken() (string, error) {
	return "mock-tuya-token", nil
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProvider := &MockTokenProvider{}

	// Setup JWT Secret for testing
	os.Setenv("JWT_SECRET", "test-secret")
	// Ensure config is loaded or secret is available to utils
	// In utils/jwt.go, it likely uses os.Getenv("JWT_SECRET") or config.
	// We might need to ensure utils picks it up. Config loading usually handles this.
	// As utils.GenerateToken uses config, let's verify utils/jwt.go usage in next thought if this fails.
	// Assuming utils.GenerateToken works with env var or we need to init config.
	
	// Pre-generate a valid token
	validToken, _ := utils.GenerateToken("user-123")

	t.Run("Valid Bearer Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		var capturedToken string
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/test", func(c *gin.Context) {
			token, exists := c.Get("access_token")
			if exists {
				capturedToken = token.(string)
			}
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if capturedToken != "mock-tuya-token" { // Expecting the token from MockTokenProvider
			t.Errorf("Expected token 'mock-tuya-token', got '%s'", capturedToken)
		}
	})

	t.Run("Token Without Bearer Prefix", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		var capturedToken string
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/test", func(c *gin.Context) {
			token, exists := c.Get("access_token")
			if exists {
				capturedToken = token.(string)
			}
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", validToken)
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if capturedToken != "mock-tuya-token" {
			t.Errorf("Expected token 'mock-tuya-token', got '%s'", capturedToken)
		}
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(AuthMiddleware(mockProvider))
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

		r.Use(AuthMiddleware(mockProvider))
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
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/test", func(c *gin.Context) {
			uid, exists := c.Get("tuya_uid")
			if exists {
				capturedUID = uid.(string)
			}
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
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
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/test", func(c *gin.Context) {
			_, uidExists = c.Get("tuya_uid")
			c.String(http.StatusOK, "success")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
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
