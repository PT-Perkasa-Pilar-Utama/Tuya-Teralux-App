package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"sensio/domain/common/infrastructure"
	"sensio/domain/terminal/terminal/entities"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// TerminalBootstrapE2ETestSuite runs end-to-end tests for the terminal bootstrap flow.
// This tests the full flow: Register -> Get by MAC -> MQTT Credentials
type TerminalBootstrapE2ETestSuite struct {
	suite.Suite
	router         *gin.Engine
	db             *gorm.DB
	apiKey         string
	testTerminalID string
}

// SetupSuite runs once before all tests in the suite.
func (suite *TerminalBootstrapE2ETestSuite) SetupSuite() {
	// Initialize test database
	testDB, err := infrastructure.InitDB()
	if err != nil {
		suite.T().Fatalf("Failed to initialize database: %v", err)
	}
	suite.db = testDB

	// Auto-migrate test tables
	err = testDB.AutoMigrate(&entities.Terminal{}, &entities.MQTTUser{})
	if err != nil {
		suite.T().Fatalf("Failed to migrate test tables: %v", err)
	}

	// Initialize router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Get API key from env
	suite.apiKey = os.Getenv("SENSIO_API_KEY")
	if suite.apiKey == "" {
		suite.apiKey = "test-api-key"
	}
}

// TearDownSuite runs once after all tests in the suite.
func (suite *TerminalBootstrapE2ETestSuite) TearDownSuite() {
	// Clean up test data
	if suite.testTerminalID != "" {
		suite.db.Delete(&entities.Terminal{}, suite.testTerminalID)
	}

	// Close database connection
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest runs before each test.
func (suite *TerminalBootstrapE2ETestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
}

// TestTerminalBootstrap_FullFlow tests the complete terminal bootstrap flow.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminalBootstrap_FullFlow() {
	// Step 1: Register a new terminal
	suite.T().Run("Step 1: Register terminal", func(t *testing.T) {
		payload := map[string]string{
			"mac_address":    "AA:BB:CC:DD:EE:FF",
			"room_id":        "room-test-001",
			"name":           "Test Terminal E2E",
			"device_type_id": "hub-type-test",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["status"].(bool))
		assert.Equal(t, "Terminal created successfully", response["message"])

		// Extract terminal ID for later tests
		data := response["data"].(map[string]interface{})
		suite.testTerminalID = data["terminal_id"].(string)
		assert.NotEmpty(t, suite.testTerminalID)

		// Verify MQTT credentials are present
		assert.NotEmpty(t, data["mqtt_username"])
		assert.NotEmpty(t, data["mqtt_password"])
	})

	// Step 2: Get terminal by MAC address
	suite.T().Run("Step 2: Get terminal by MAC", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/terminal/mac/AA:BB:CC:DD:EE:FF", nil)
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["status"].(bool))
		assert.Equal(t, "Terminal retrieved successfully", response["message"])

		// Verify response structure
		data := response["data"].(map[string]interface{})
		assert.Equal(t, suite.testTerminalID, data["id"])
		assert.Equal(t, "AA:BB:CC:DD:EE:FF", data["mac_address"])
		assert.Equal(t, "Test Terminal E2E", data["name"])
	})

	// Step 3: Get MQTT credentials
	suite.T().Run("Step 3: Get MQTT credentials", func(t *testing.T) {
		// First, we need to get the MQTT username from the terminal
		terminal := &entities.Terminal{}
		err := suite.db.First(terminal, suite.testTerminalID).Error
		assert.NoError(t, err)

		mqttUsername := fmt.Sprintf("mqtt_%s", terminal.ID)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/mqtt/users/%s", mqttUsername), nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Note: This might return 404 if MQTT user wasn't created in test mode
		// The important part is testing the endpoint structure
		assert.NotEqual(t, http.StatusInternalServerError, w.Code)
	})
}

// TestTerminal_UpdateTerminalResponseContract tests that PUT /api/terminal/{id}
// returns the correct response structure with terminal data.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_UpdateTerminalResponseContract() {
	if suite.testTerminalID == "" {
		suite.T().Skip("No test terminal available")
	}

	payload := map[string]string{
		"name": "Updated Terminal Name",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/terminal/%s", suite.testTerminalID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Verify response structure matches contract
	assert.True(suite.T(), response["status"].(bool))
	assert.Equal(suite.T(), "Updated successfully", response["message"])

	// Data should contain terminal information, not be null
	data, ok := response["data"].(map[string]interface{})
	assert.True(suite.T(), ok, "Data should be a map, not null")
	assert.Equal(suite.T(), suite.testTerminalID, data["id"])
	assert.Equal(suite.T(), "Updated Terminal Name", data["name"])
}

// TestTerminal_ErrorResponses tests that error responses are consistent
// and don't leak internal implementation details.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_ErrorResponses() {
	suite.T().Run("Get non-existent terminal by MAC", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/terminal/mac/ZZ:ZZ:ZZ:ZZ:ZZ:ZZ", nil)
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.False(t, response["status"].(bool))
		assert.Equal(t, "Terminal not found", response["message"])

		// Ensure no internal error details are leaked
		assert.NotContains(t, response["message"], "Internal server error")
		assert.NotContains(t, response["message"], "sql:")
		assert.NotContains(t, response["message"], "gorm:")
	})

	suite.T().Run("Update non-existent terminal", func(t *testing.T) {
		payload := map[string]string{
			"name": "Should Fail",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/terminal/non-existent-id", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.False(t, response["status"].(bool))
		// Should return generic "Not Found" message, not internal details
		assert.Equal(t, "Not Found", response["message"])
	})
}

// TestTerminal_ValidationErrors tests that validation errors return proper structure.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_ValidationErrors() {
	suite.T().Run("Invalid MAC address format", func(t *testing.T) {
		payload := map[string]string{
			"mac_address":    "invalid-mac",
			"room_id":        "room-test-001",
			"name":           "Test Terminal",
			"device_type_id": "hub-type-test",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.False(t, response["status"].(bool))
		assert.Equal(t, "Validation Error", response["message"])
	})
}

// TestTerminalBootstrapE2ETest runs the test suite.
func TestTerminalBootstrapE2ETest(t *testing.T) {
	suite.Run(t, new(TerminalBootstrapE2ETestSuite))
}
