package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testRouter *gin.Engine
	testAPIKey string
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	router := gin.New()
	testAPIKey = os.Getenv("SENSIO_API_KEY")
	if testAPIKey == "" {
		testAPIKey = "test-api-key"
	}
	return router
}

func TestMain(m *testing.M) {
	testRouter = setupTestRouter()
	os.Exit(m.Run())
}

func makeRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func assertResponseOK(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["status"].(bool), "Response status should be true")
	return response
}

func assertResponseCreated(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["status"].(bool), "Response status should be true")
	return response
}

func assertResponseNotFound(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["status"].(bool), "Response status should be false")
	return response
}

func assertResponseBadRequest(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}

func assertResponseUnauthorized(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func assertResponseError(t *testing.T, w *httptest.ResponseRecorder, statusCode int) map[string]interface{} {
	assert.Equal(t, statusCode, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}

func getData(response map[string]interface{}) map[string]interface{} {
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil
	}
	return data
}
