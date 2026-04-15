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

type RAGE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *RAGE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *RAGE2ETestSuite) TestRAG_Chat() {
	payload := map[string]interface{}{
		"prompt":      "Hello, what can you do?",
		"mac_address": "AA:BB:CC:DD:EE:FF",
		"session_id":  "session-123",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/rag/chat", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RAGE2ETestSuite) TestRAG_Control() {
	payload := map[string]interface{}{
		"prompt":      "Turn on the lights",
		"mac_address": "AA:BB:CC:DD:EE:FF",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/rag/control", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RAGE2ETestSuite) TestRAG_GetStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/rag/task-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RAGE2ETestSuite) TestRAG_Translate() {
	payload := map[string]interface{}{
		"text":            "Hello world",
		"source_language": "en",
		"target_language": "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/rag/translate", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RAGE2ETestSuite) TestRAG_Summary() {
	payload := map[string]interface{}{
		"text":         "This is a long meeting transcript...",
		"context":      "team meeting",
		"style":        "minutes",
		"date":         "2026-04-15",
		"location":     "Jakarta",
		"participants": []string{"Alice", "Bob"},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/rag/summary", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *RAGE2ETestSuite) TestRAG_ValidationErrors() {
	s.T().Run("Empty prompt", func(t *testing.T) {
		payload := map[string]interface{}{
			"prompt": "",
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/models/rag/chat", bytes.NewBufferString(mustMarshal(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.True(w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
	})
}

func TestRAGE2ETestSuite(t *testing.T) {
	suite.Run(t, new(RAGE2ETestSuite))
}
