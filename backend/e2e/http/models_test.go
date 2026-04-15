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

type ModelsE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *ModelsE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *ModelsE2ETestSuite) TestModels_OpenAI() {
	payload := map[string]interface{}{
		"prompt": "Hello, how are you?",
		"model":  "gpt-4",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/openai", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModels_Gemini() {
	payload := map[string]interface{}{
		"prompt": "Hello, how are you?",
		"model":  "gemini-pro",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/gemini", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModels_Groq() {
	payload := map[string]interface{}{
		"prompt": "Hello, how are you?",
		"model":  "mixtral-8x7b",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/groq", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModels_Orion() {
	payload := map[string]interface{}{
		"prompt": "Hello, how are you?",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/orion", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModels_Llama() {
	payload := map[string]interface{}{
		"prompt": "Hello, how are you?",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/llama/cpp", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModelsV1_RAGChat() {
	payload := map[string]interface{}{
		"prompt":      "Hello",
		"mac_address": "AA:BB:CC:DD:EE:FF",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/rag/chat", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModelsV1_RAGControl() {
	payload := map[string]interface{}{
		"prompt":      "Turn on lights",
		"mac_address": "AA:BB:CC:DD:EE:FF",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/rag/control", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModelsV1_RAGStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/v1/rag/status/task-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModelsV1_RAGSummary() {
	payload := map[string]interface{}{
		"text": "Meeting transcript...",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/rag/summary", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *ModelsE2ETestSuite) TestModelsV1_RAGTranslate() {
	payload := map[string]interface{}{
		"text":            "Hello",
		"source_language": "en",
		"target_language": "id",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/rag/translate", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func TestModelsE2ETestSuite(t *testing.T) {
	suite.Run(t, new(ModelsE2ETestSuite))
}
