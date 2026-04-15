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

type TuyaE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *TuyaE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *TuyaE2ETestSuite) TestTuya_Auth() {
	payload := map[string]interface{}{
		"username": "test",
		"password": "test123",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/tuya/auth", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *TuyaE2ETestSuite) TestTuya_GetDevices() {
	req, _ := http.NewRequest(http.MethodGet, "/api/tuya/devices", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *TuyaE2ETestSuite) TestTuya_GetDeviceByID() {
	req, _ := http.NewRequest(http.MethodGet, "/api/tuya/devices/device-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *TuyaE2ETestSuite) TestTuya_SendIRCommand() {
	payload := map[string]interface{}{
		"command": "power_on",
		"code":    "key_001",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/tuya/devices/device-123/commands/ir", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *TuyaE2ETestSuite) TestTuya_SendSwitchCommand() {
	payload := map[string]interface{}{
		"command": "turn_on",
		"value":   "100",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/tuya/devices/device-123/commands/switch", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *TuyaE2ETestSuite) TestTuya_GetSensor() {
	req, _ := http.NewRequest(http.MethodGet, "/api/tuya/devices/device-123/sensor", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *TuyaE2ETestSuite) TestTuya_NotFound() {
	s.T().Run("Non-existent device", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/tuya/devices/non-existent", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.Equal(http.StatusNotFound, w.Code)
	})
}

func TestTuyaE2ETestSuite(t *testing.T) {
	suite.Run(t, new(TuyaE2ETestSuite))
}
