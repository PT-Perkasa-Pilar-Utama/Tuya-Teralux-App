package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTuyaErrorMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Intercepts code: 1010 in response body", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(TuyaErrorMiddleware())
		r.GET("/test", func(c *gin.Context) {
			// Simulate response that contains "code: 1010"
			c.String(http.StatusOK, `{"success":false,"code: 1010","msg":"token invalid"}`)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		// Should be intercepted and converted to 401
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		// Check response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response["status"] != false {
			t.Error("Expected status to be false")
		}

		if msg, ok := response["message"].(string); !ok || msg != "Token expired. Please login or refresh the token" {
			t.Errorf("Expected specific error message, got: %v", response["message"])
		}
	})

	t.Run("Passes through normal responses", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(TuyaErrorMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    "test data",
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		// Should pass through as-is
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Response should contain original data
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if response["success"] != true {
			t.Error("Expected success to be true")
		}
	})

	t.Run("Passes through other error codes", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(TuyaErrorMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusBadRequest, `{"code":400,"msg":"bad request"}`)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		// Should pass through as-is (not intercepted)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
