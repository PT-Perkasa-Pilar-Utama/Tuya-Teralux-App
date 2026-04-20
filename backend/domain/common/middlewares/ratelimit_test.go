package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Health endpoint is excluded from rate limiting", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(RateLimitMiddleware())
		r.GET("/api/health", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Request under limit passes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		config := RateLimitConfig{
			RequestsPerMinute: 100,
			burst:             100,
		}
		r.Use(RateLimitMiddlewareWithConfig(config))
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Rate limit headers are set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		config := RateLimitConfig{
			RequestsPerMinute: 100,
			burst:             100,
		}
		r.Use(RateLimitMiddlewareWithConfig(config))
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		c.Request = req

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		limit := w.Header().Get("X-RateLimit-Limit")
		if limit != "100" {
			t.Errorf("Expected X-RateLimit-Limit 100, got %s", limit)
		}
	})

	t.Run("Request over limit returns 429", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		config := RateLimitConfig{
			RequestsPerMinute: 1,
			burst:             1,
		}
		r.Use(RateLimitMiddlewareWithConfig(config))
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.3:12345"
		c.Request = req

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("First request: expected status 200, got %d", w.Code)
		}

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c2.Request.RemoteAddr = "192.168.1.3:12345"

		r.ServeHTTP(w2, c2.Request)

		if w2.Code != http.StatusTooManyRequests {
			t.Errorf("Second request: expected status 429, got %d", w2.Code)
		}

		retryAfter := w2.Header().Get("Retry-After")
		if retryAfter == "" {
			t.Error("Expected Retry-After header to be set")
		}
	})

	t.Run("Different IPs have separate limits", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		config := RateLimitConfig{
			RequestsPerMinute: 1,
			burst:             1,
		}
		r.Use(RateLimitMiddlewareWithConfig(config))
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.10:12345"
		c.Request = req1

		r.ServeHTTP(w, req1)
		if w.Code != http.StatusOK {
			t.Errorf("First IP first request: expected status 200, got %d", w.Code)
		}

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.10:12345"
		c2.Request = req2

		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusTooManyRequests {
			t.Errorf("First IP second request: expected status 429, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req3.RemoteAddr = "192.168.1.11:12345"
		c3.Request = req3

		r.ServeHTTP(w3, req3)
		if w3.Code != http.StatusOK {
			t.Errorf("Second IP: expected status 200, got %d", w3.Code)
		}
	})

	t.Run("Response body contains rate limit message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		config := RateLimitConfig{
			RequestsPerMinute: 1,
			burst:             1,
		}
		r.Use(RateLimitMiddlewareWithConfig(config))
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.20:12345"
		c.Request = req

		r.ServeHTTP(w, req)

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.20:12345"
		c2.Request = req2

		r.ServeHTTP(w2, req2)

		body := w2.Body.String()
		if !strings.Contains(body, "Too Many Requests") {
			t.Errorf("Expected body to contain 'Too Many Requests', got: %s", body)
		}
	})
}

func TestRateLimitersCleanup(t *testing.T) {
	limiters := newRateLimiters(RateLimitConfig{RequestsPerMinute: 10, burst: 10})

	limiters.getLimiter("192.168.1.1")
	limiters.getLimiter("192.168.1.2")

	if len(limiters.limiters) != 2 {
		t.Errorf("Expected 2 limiters, got %d", len(limiters.limiters))
	}

	limiters.cleanupStale(0)

	if len(limiters.limiters) != 0 {
		t.Errorf("Expected 0 limiters after cleanup, got %d", len(limiters.limiters))
	}
}
