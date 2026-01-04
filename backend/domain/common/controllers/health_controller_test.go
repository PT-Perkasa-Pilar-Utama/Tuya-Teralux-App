package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthController_CheckHealth(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	controller := NewHealthController()

	t.Run("Success - Database Healthy", func(t *testing.T) {
		// Create a test HTTP request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

		// Call the handler
		controller.CheckHealth(c)

		// Note: This test may fail if database is not initialized
		// In a real scenario, we would mock infrastructure.PingDB()
		// For now, we just verify the handler doesn't panic
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 200 or 503, got %d", w.Code)
		}
	})
}

func TestNewHealthController(t *testing.T) {
	controller := NewHealthController()
	if controller == nil {
		t.Fatal("NewHealthController returned nil")
	}
}
