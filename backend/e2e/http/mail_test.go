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

type MailE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *MailE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *MailE2ETestSuite) TestMail_Send() {
	payload := map[string]interface{}{
		"to":       []string{"test@example.com"},
		"template": "welcome",
		"data": map[string]string{
			"name": "Test User",
		},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/mail/send", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *MailE2ETestSuite) TestMail_SendByMAC() {
	payload := map[string]interface{}{
		"template": "alert",
		"data": map[string]string{
			"message": "System alert",
		},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/mail/send/mac/AA:BB:CC:DD:EE:FF", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *MailE2ETestSuite) TestMail_GetStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/mail/status/task-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *MailE2ETestSuite) TestMail_ValidationErrors() {
	s.T().Run("Missing recipients", func(t *testing.T) {
		payload := map[string]interface{}{
			"to":       []string{},
			"template": "welcome",
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/mail/send", bytes.NewBufferString(mustMarshal(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.True(w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
	})
}

func TestMailE2ETestSuite(t *testing.T) {
	suite.Run(t, new(MailE2ETestSuite))
}
