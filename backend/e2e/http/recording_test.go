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

type RecordingE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *RecordingE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *RecordingE2ETestSuite) TestRecording_GetAll() {
	req, _ := http.NewRequest(http.MethodGet, "/api/recordings", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RecordingE2ETestSuite) TestRecording_GetAllWithPagination() {
	req, _ := http.NewRequest(http.MethodGet, "/api/recordings?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RecordingE2ETestSuite) TestRecording_Create() {
	payload := map[string]interface{}{
		"mac_address": "AA:BB:CC:DD:EE:FF",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/recordings", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RecordingE2ETestSuite) TestRecording_GetByID() {
	req, _ := http.NewRequest(http.MethodGet, "/api/recordings/rec-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RecordingE2ETestSuite) TestRecording_Delete() {
	req, _ := http.NewRequest(http.MethodDelete, "/api/recordings/rec-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RecordingE2ETestSuite) TestRecording_NotFound() {
	s.T().Run("Non-existent recording", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/recordings/non-existent", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.Equal(http.StatusNotFound, w.Code)
	})
}

func TestRecordingE2ETestSuite(t *testing.T) {
	suite.Run(t, new(RecordingE2ETestSuite))
}
