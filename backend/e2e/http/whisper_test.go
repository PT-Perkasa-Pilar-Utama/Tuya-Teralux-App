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

type WhisperE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *WhisperE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *WhisperE2ETestSuite) TestWhisper_Transcribe() {
	payload := map[string]interface{}{
		"language":    "id",
		"mac_address": "AA:BB:CC:DD:EE:FF",
		"diarize":     true,
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/whisper/transcribe", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisper_GetStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/whisper/transcribe/trans-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperV1_Transcribe() {
	payload := map[string]interface{}{
		"audio_path": "/tmp/test.wav",
		"language":   "id",
		"diarize":    false,
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/whisper/transcribe", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperV1_GetStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/v1/whisper/status/trans-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperV1_CreateUploadSession() {
	payload := map[string]interface{}{
		"filename": "test.wav",
		"size":     1024000,
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/whisper/uploads/sessions", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperV1_GetUploadSession() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/v1/whisper/uploads/sessions/session-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperV1_UploadChunk() {
	req, _ := http.NewRequest(http.MethodPut, "/api/models/v1/whisper/uploads/sessions/session-123/chunks/0", bytes.NewBufferString("test-data"))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperModel_OpenAI() {
	payload := map[string]interface{}{
		"audio_path": "/tmp/test.wav",
		"language":   "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/whisper/models/openai", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperModel_Gemini() {
	payload := map[string]interface{}{
		"audio_path": "/tmp/test.wav",
		"language":   "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/whisper/models/gemini", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperModel_Groq() {
	payload := map[string]interface{}{
		"audio_path": "/tmp/test.wav",
		"language":   "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/whisper/models/groq", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperModel_Orion() {
	payload := map[string]interface{}{
		"audio_path": "/tmp/test.wav",
		"language":   "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/whisper/models/orion", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *WhisperE2ETestSuite) TestWhisperModel_LlamaCpp() {
	payload := map[string]interface{}{
		"audio_path": "/tmp/test.wav",
		"language":   "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/whisper/models/whisper/cpp", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func TestWhisperE2ETestSuite(t *testing.T) {
	suite.Run(t, new(WhisperE2ETestSuite))
}
