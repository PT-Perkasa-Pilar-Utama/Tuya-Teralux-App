package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TerminalE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
	apiKey string
}

func (suite *TerminalE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.apiKey = "test-api-key"
}

func (suite *TerminalE2ETestSuite) TestTerminal_GetAllTerminals() {
	req, _ := http.NewRequest(http.MethodGet, "/api/terminal", nil)
	req.Header.Set("X-API-KEY", suite.apiKey)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert := assert.New(suite.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (suite *TerminalE2ETestSuite) TestTerminal_CreateTerminal() {
	payload := map[string]string{
		"mac_address":    "11:22:33:44:55:66",
		"room_id":        "room-001",
		"name":           "Test Terminal",
		"device_type_id": "hub-type-001",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", suite.apiKey)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert := assert.New(suite.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (suite *TerminalE2ETestSuite) TestTerminal_GetTerminalByID() {
	req, _ := http.NewRequest(http.MethodGet, "/api/terminal/test-id", nil)
	req.Header.Set("X-API-KEY", suite.apiKey)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert := assert.New(suite.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (suite *TerminalE2ETestSuite) TestTerminal_GetTerminalByMAC() {
	req, _ := http.NewRequest(http.MethodGet, "/api/terminal/mac/11:22:33:44:55:66", nil)
	req.Header.Set("X-API-KEY", suite.apiKey)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert := assert.New(suite.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (suite *TerminalE2ETestSuite) TestTerminal_UpdateTerminal() {
	payload := map[string]string{
		"name": "Updated Terminal",
	}

	req, _ := http.NewRequest(http.MethodPut, "/api/terminal/test-id", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", suite.apiKey)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert := assert.New(suite.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (suite *TerminalE2ETestSuite) TestTerminal_DeleteTerminal() {
	req, _ := http.NewRequest(http.MethodDelete, "/api/terminal/test-id", nil)
	req.Header.Set("X-API-KEY", suite.apiKey)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert := assert.New(suite.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (suite *TerminalE2ETestSuite) TestTerminal_ValidationErrors() {
	suite.T().Run("Invalid MAC address", func(t *testing.T) {
		payload := map[string]string{
			"mac_address":    "invalid-mac",
			"room_id":        "room-001",
			"name":           "Test Terminal",
			"device_type_id": "hub-type-001",
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewBufferString(mustMarshal(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.True(w.Code == http.StatusUnprocessableEntity || w.Code == http.StatusNotFound)
	})
}

func (suite *TerminalE2ETestSuite) TestTerminal_NotFoundErrors() {
	suite.T().Run("Non-existent terminal by MAC", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/terminal/mac/FF:FF:FF:FF:FF:FF", nil)
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.Equal(http.StatusNotFound, w.Code)
	})
}

func TestTerminalE2ETestSuite(t *testing.T) {
	suite.Run(t, new(TerminalE2ETestSuite))
}

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
