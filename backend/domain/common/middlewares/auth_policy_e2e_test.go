package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"sensio/domain/common/utils"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthPolicyE2E(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("GO_TEST", "true")
	defer func() { _ = os.Unsetenv("GO_TEST") }()
	_ = os.Setenv("JWT_SECRET", "test-jwt-secret")
	_ = os.Setenv("API_KEY", "test-api-key-123")
	defer func() {
		_ = os.Unsetenv("JWT_SECRET")
		_ = os.Unsetenv("API_KEY")
	}()
	utils.AppConfig = nil

	validBearerToken, err := utils.GenerateToken("test-user-123")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Test 1: Bootstrap route with valid API key - ALLOWED
	t.Run("Bootstrap POST /api/terminal with valid API key - ALLOWED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(ApiKeyMiddleware())
		r.POST("/api/terminal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "created"})
		})
		reqBody := `{"name": "test-terminal", "mac_address": "AA:BB:CC:DD:EE:FF"}`
		req := httptest.NewRequest(http.MethodPost, "/api/terminal", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", "test-api-key-123")
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid API key on bootstrap route, got %d", w.Code)
		}
	})

	// Test 2: Bootstrap route without API key - DENIED
	t.Run("Bootstrap POST /api/terminal without API key - DENIED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(ApiKeyMiddleware())
		r.POST("/api/terminal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "created"})
		})
		reqBody := `{"name": "test-terminal", "mac_address": "AA:BB:CC:DD:EE:FF"}`
		req := httptest.NewRequest(http.MethodPost, "/api/terminal", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for missing API key, got %d", w.Code)
		}
	})

	// Test 3: Bootstrap GET /api/terminal/mac/:mac with valid API key - ALLOWED
	t.Run("Bootstrap GET /api/terminal/mac/:mac with valid API key - ALLOWED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(ApiKeyMiddleware())
		r.GET("/api/terminal/mac/:mac", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"mac": c.Param("mac")})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/terminal/mac/AA:BB:CC:DD:EE:FF", nil)
		req.Header.Set("X-API-KEY", "test-api-key-123")
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid API key on MAC lookup, got %d", w.Code)
		}
	})

	// Test 4: Bootstrap GET /api/tuya/auth with valid API key - ALLOWED
	t.Run("Bootstrap GET /api/tuya/auth with valid API key - ALLOWED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(ApiKeyMiddleware())
		r.GET("/api/tuya/auth", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"token": "tuya-token"})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/tuya/auth", nil)
		req.Header.Set("X-API-KEY", "test-api-key-123")
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid API key on Tuya auth, got %d", w.Code)
		}
	})

	// Test 5: Operational route with valid Bearer - ALLOWED
	t.Run("Operational GET /api/terminal with valid Bearer - ALLOWED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		mockProvider := &MockTokenProvider{}
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/api/terminal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"terminals": []interface{}{}})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/terminal", nil)
		req.Header.Set("Authorization", "Bearer "+validBearerToken)
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid Bearer on operational route, got %d", w.Code)
		}
	})

	// Test 6: Operational route without Bearer - DENIED
	t.Run("Operational GET /api/terminal without Bearer - DENIED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		mockProvider := &MockTokenProvider{}
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/api/terminal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"terminals": []interface{}{}})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/terminal", nil)
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for missing Bearer, got %d", w.Code)
		}
	})

	// Test 7: Operational route with API key only - DENIED
	t.Run("Operational GET /api/terminal with API key only - DENIED", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		mockProvider := &MockTokenProvider{}
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/api/terminal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"terminals": []interface{}{}})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/terminal", nil)
		req.Header.Set("X-API-KEY", "test-api-key-123")
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for API key only on operational route, got %d", w.Code)
		}
	})

	// Test 8: Expired Bearer Token - DENIED
	t.Run("Expired Bearer Token - DENIED", func(t *testing.T) {
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJ0ZXN0LXVzZXIiLCJleHAiOjAsImlhdCI6MH0.invalid"
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		mockProvider := &MockTokenProvider{}
		r.Use(AuthMiddleware(mockProvider))
		r.GET("/api/terminal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"terminals": []interface{}{}})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/terminal", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		c.Request = req
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for expired token, got %d", w.Code)
		}
	})
}
