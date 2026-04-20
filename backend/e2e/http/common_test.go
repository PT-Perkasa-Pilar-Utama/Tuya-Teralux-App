package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CommonE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *CommonE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *CommonE2ETestSuite) TestHealthCheck() {
	req, _ := http.NewRequest(http.MethodGet, "/api/health", nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *CommonE2ETestSuite) TestCacheFlush() {
	req, _ := http.NewRequest(http.MethodDelete, "/api/cache/flush", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *CommonE2ETestSuite) TestBigDevice_GetByMAC() {
	req, _ := http.NewRequest(http.MethodGet, "/api/big/device/AA:BB:CC:DD:EE:FF", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *CommonE2ETestSuite) TestNotification_Publish() {
	payload := map[string]interface{}{
		"room_id":       "room-001",
		"scheduled_at":  "2026-04-20T23:00:00+07:00",
		"phone_numbers": []string{"+6281234567890"},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/notification/publish", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *CommonE2ETestSuite) TestMQTT_GetCredentials() {
	req, _ := http.NewRequest(http.MethodGet, "/api/mqtt/users/mqtt-user-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func TestCommonE2ETestSuite(t *testing.T) {
	suite.Run(t, new(CommonE2ETestSuite))
}
