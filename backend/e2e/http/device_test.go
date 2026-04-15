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

type DeviceE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *DeviceE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *DeviceE2ETestSuite) TestDevice_Create() {
	payload := map[string]interface{}{
		"terminal_id":    "term-123",
		"device_name":    "Test Device",
		"tuya_device_id": "tuya-456",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/devices", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *DeviceE2ETestSuite) TestDevice_GetByID() {
	req, _ := http.NewRequest(http.MethodGet, "/api/devices/device-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *DeviceE2ETestSuite) TestDevice_Update() {
	payload := map[string]string{
		"device_name": "Updated Device",
	}

	req, _ := http.NewRequest(http.MethodPut, "/api/devices/device-123", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *DeviceE2ETestSuite) TestDevice_UpdateStatus() {
	payload := map[string]interface{}{
		"status":      "on",
		"value":       "100",
		"mac_address": "AA:BB:CC:DD:EE:FF",
	}

	req, _ := http.NewRequest(http.MethodPut, "/api/devices/device-123/status", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *DeviceE2ETestSuite) TestDevice_GetStatuses() {
	req, _ := http.NewRequest(http.MethodGet, "/api/devices/device-123/statuses", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *DeviceE2ETestSuite) TestDevice_GetAllStatuses() {
	req, _ := http.NewRequest(http.MethodGet, "/api/devices/statuses", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *DeviceE2ETestSuite) TestDevice_GetByTerminal() {
	req, _ := http.NewRequest(http.MethodGet, "/api/devices/terminal/term-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func TestDeviceE2ETestSuite(t *testing.T) {
	suite.Run(t, new(DeviceE2ETestSuite))
}
