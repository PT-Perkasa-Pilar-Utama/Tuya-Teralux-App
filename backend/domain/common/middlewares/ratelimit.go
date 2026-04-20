package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerMinute int
	burst             int
}

// DefaultRateLimitConfig returns the default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 100,
		burst:             100,
	}
}

// ipLimiter holds a per-IP rate limiter.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiters manages per-IP rate limiters with thread-safe access.
type RateLimiters struct {
	limiters map[string]*ipLimiter
	mu       sync.Mutex
	config   RateLimitConfig
}

func newRateLimiters(config RateLimitConfig) *RateLimiters {
	return &RateLimiters{
		limiters: make(map[string]*ipLimiter),
		config:   config,
	}
}

func (r *RateLimiters) getLimiter(ip string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, exists := r.limiters[ip]
	if !exists {
		burst := r.config.burst
		if burst <= 0 {
			burst = r.config.RequestsPerMinute
		}
		limiter = &ipLimiter{
			limiter:  rate.NewLimiter(rate.Limit(float64(r.config.RequestsPerMinute)/60.0), burst),
			lastSeen: time.Now(),
		}
		r.limiters[ip] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter
}

// cleanupStale removes limiters that haven't been seen for the specified duration.
func (r *RateLimiters) cleanupStale(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for ip, limiter := range r.limiters {
		if now.Sub(limiter.lastSeen) > maxAge {
			delete(r.limiters, ip)
		}
	}
}

var (
	globalRateLimiters *RateLimiters
	once               sync.Once
)

// getRateLimiters returns the global rate limiters instance.
func getRateLimiters() *RateLimiters {
	once.Do(func() {
		globalRateLimiters = newRateLimiters(DefaultRateLimitConfig())
	})
	return globalRateLimiters
}

// RateLimitMiddleware returns a Gin middleware that implements IP-based rate limiting.
// Uses token bucket algorithm with 100 requests per minute per IP.
// Returns 429 Too Many Requests with Retry-After header when limit is exceeded.
func RateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddlewareWithConfig(DefaultRateLimitConfig())
}

// RateLimitMiddlewareWithConfig returns a Gin middleware with custom rate limit configuration.
func RateLimitMiddlewareWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiters := newRateLimiters(config)

	// Start cleanup goroutine to prevent memory leaks
	stopCleanup := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limiters.cleanupStale(10 * time.Minute)
			case <-stopCleanup:
				return
			}
		}
	}()

	return func(c *gin.Context) {
		// Skip rate limiting for health check endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/health") {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := limiters.getLimiter(ip)

		if !limiter.Allow() {
			retryAfter := strconv.Itoa(int(time.Minute / time.Second))
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", retryAfter)

			utils.LogWarn("RateLimitMiddleware: rate limit exceeded for IP %s", ip)
			c.JSON(http.StatusTooManyRequests, dtos.StandardResponse{
				Status:  false,
				Message: "Too Many Requests. Please retry after " + retryAfter + " seconds.",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", "0")

		c.Next()
	}
}
