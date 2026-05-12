package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	controllers "sensio/domain/terminal/terminal/controllers"
	"sensio/domain/terminal/terminal/entities"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	terminal_services "sensio/domain/terminal/terminal/services"
	usecases "sensio/domain/terminal/terminal/usecases"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TerminalBootstrapE2ETestSuite struct {
	suite.Suite
	router         *gin.Engine
	db             *gorm.DB
	badger         *infrastructure.BadgerService
	apiKey         string
	bearerToken    string
	testTerminalID string
	testMacAddress string
}

func (suite *TerminalBootstrapE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	utils.LoadConfig()

	cfg := utils.GetConfig()
	suite.apiKey = os.Getenv("API_KEY")
	if suite.apiKey == "" {
		suite.apiKey = "test-api-key-123"
		_ = os.Setenv("API_KEY", suite.apiKey)
		utils.LoadConfig()
		cfg = utils.GetConfig()
	}

	seed := time.Now().UnixNano()
	suite.testMacAddress = fmt.Sprintf("02:00:%02X:%02X:%02X:%02X", byte(seed>>24), byte(seed>>16), byte(seed>>8), byte(seed))
	if cfg.JWTSecret != "" {
		claims := jwt.MapClaims{
			"uid": "test-user-uid",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err == nil {
			suite.bearerToken = "Bearer " + tokenString
		} else {
			suite.bearerToken = "Bearer invalid-token"
		}
	} else {
		suite.bearerToken = "Bearer invalid-token"
	}

	testDB, err := infrastructure.InitDB()
	if err != nil {
		suite.T().Fatalf("Failed to initialize database: %v", err)
	}
	suite.db = testDB

	err = testDB.AutoMigrate(&entities.Terminal{}, &entities.MQTTUser{})
	if err != nil {
		suite.T().Fatalf("Failed to migrate test tables: %v", err)
	}

	suite.router = gin.New()
	suite.router.Use(gin.Recovery())

	publicGroup := suite.router.Group("/")
	publicGroup.Use(func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		cfg := utils.GetConfig()
		if cfg.ApiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
			c.Abort()
			return
		}
		if apiKey != cfg.ApiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid API Key"})
			c.Abort()
			return
		}
		c.Next()
	})

	// Build terminal controllers manually (no full module to avoid Tuya deps)
	badger, err := infrastructure.NewBadgerService("./tmp/badger_e2e_test")
	if err != nil {
		suite.T().Fatalf("Failed to init BadgerService: %v", err)
	}
	suite.badger = badger

	// Build MQTT client with real or stub config
	mqttCfg := utils.GetConfig()
	mqttAuthClient := terminal_services.NewMqttAuthClient(mqttCfg.EmqxAuthBaseURL, mqttCfg.EmqxAuthApiKey)

	// Terminal repository
	terminalRepo := terminal_repositories.NewTerminalRepository(badger)

	// External service (mock-able for e2e)
	externalService := terminal_services.NewMacRegistrationExternalService()

	// Use cases
	createUC := usecases.NewCreateTerminalUseCase(terminalRepo, externalService, mqttAuthClient)
	getByMACUC := usecases.NewGetTerminalByMACUseCase(terminalRepo, mqttAuthClient)
	getAllUC := usecases.NewGetAllTerminalUseCase(terminalRepo)
	getByIDUC := usecases.NewGetTerminalByIDUseCase(terminalRepo, nil)
	updateUC := usecases.NewUpdateTerminalUseCase(terminalRepo)
	deleteUC := usecases.NewDeleteTerminalUseCase(terminalRepo, mqttAuthClient)

	// Controllers
	createCtrl := controllers.NewCreateTerminalController(createUC)
	getByMACCtrl := controllers.NewGetTerminalByMACController(getByMACUC)
	getAllCtrl := controllers.NewGetAllTerminalController(getAllUC)
	getByIDCtrl := controllers.NewGetTerminalByIDController(getByIDUC)
	updateCtrl := controllers.NewUpdateTerminalController(updateUC)
	deleteCtrl := controllers.NewDeleteTerminalController(deleteUC)
	mqttCredCtrl := controllers.NewGetMQTTCredentialsController(mqttAuthClient)

	// Public bootstrap routes (API-key protected)
	publicGroup.POST("/api/terminal", createCtrl.CreateTerminal)
	publicGroup.GET("/api/terminal/mac/:mac", getByMACCtrl.GetTerminalByMAC)

	// Protected group (Bearer auth)
	protectedGroup := suite.router.Group("/")
	protectedGroup.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header is required"})
			c.Abort()
			return
		}
		fields := splitAuthHeader(authHeader)
		if len(fields) != 2 || fields[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid Authorization Header"})
			c.Abort()
			return
		}
		tokenString := fields[1]
		uid, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired token"})
			c.Abort()
			return
		}
		c.Set("uid", uid)
		c.Next()
	})

	protectedGroup.GET("/api/terminal", getAllCtrl.GetAllTerminal)
	protectedGroup.GET("/api/terminal/:id", getByIDCtrl.GetTerminalByID)
	protectedGroup.PUT("/api/terminal/:id", updateCtrl.UpdateTerminal)
	protectedGroup.DELETE("/api/terminal/:id", deleteCtrl.DeleteTerminal)
	protectedGroup.GET("/api/mqtt/users/:username", mqttCredCtrl.GetMQTTCredentials)
}

// splitAuthHeader is a helper to split "Bearer token" safely.
func splitAuthHeader(auth string) []string {
	const bearerPrefix = "Bearer "
	if len(auth) > len(bearerPrefix) && auth[:7] == bearerPrefix {
		return []string{"Bearer", auth[7:]}
	}
	return []string{}
}

// TearDownSuite runs once after all tests in the suite.
func (suite *TerminalBootstrapE2ETestSuite) TearDownSuite() {
	if suite.db != nil {
		if suite.testTerminalID != "" {
			suite.db.Where("id = ?", suite.testTerminalID).Delete(&entities.Terminal{})
		}
		if suite.testMacAddress != "" {
			suite.db.Where("mac_address = ?", suite.testMacAddress).Delete(&entities.Terminal{})
		}
		sqlDB, err := suite.db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
	if suite.badger != nil {
		suite.badger.Close()
	}
}

func (suite *TerminalBootstrapE2ETestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	if suite.badger != nil && suite.testMacAddress != "" {
		_ = suite.badger.Delete(fmt.Sprintf("terminal:mac:%s", suite.testMacAddress))
	}
}

// TestTerminalBootstrap_FullFlow tests the complete terminal bootstrap flow.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminalBootstrap_FullFlow() {
	suite.T().Run("Step 1: Register terminal", func(t *testing.T) {
		payload := map[string]string{
			"mac_address":    suite.testMacAddress,
			"room_id":        "1",
			"name":           "Test Terminal E2E",
			"device_type_id": "1",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Expected 201, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["status"].(bool), "Expected status=true, got %v", response["status"])
		assert.Equal(t, "Terminal created successfully", response["message"])

		data := response["data"].(map[string]interface{})
		suite.testTerminalID = data["terminal_id"].(string)
		assert.NotEmpty(t, suite.testTerminalID)
		assert.NotEmpty(t, data["mqtt_username"])
	})

	suite.T().Run("Step 2: Get terminal by MAC", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/terminal/mac/%s", suite.testMacAddress), nil)
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected 200, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["status"].(bool))
		assert.Equal(t, "Terminal retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		terminal := data["terminal"].(map[string]interface{})
		assert.Equal(t, suite.testTerminalID, terminal["id"])
		assert.Equal(t, suite.testMacAddress, terminal["mac_address"])
		assert.Equal(t, "Test Terminal E2E", terminal["name"])
	})

	suite.T().Run("Step 3: Get MQTT credentials", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/mqtt/users/%s", suite.testMacAddress), nil)
		req.Header.Set("Authorization", suite.bearerToken)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusInternalServerError, w.Code, "Expected not 500, got %d: %s", w.Code, w.Body.String())
	})
}

// TestTerminal_UpdateTerminalResponseContract tests PUT /api/terminal/{id} response contract.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_UpdateTerminalResponseContract() {
	if suite.testTerminalID == "" {
		suite.T().Skip("No test terminal available")
	}

	payload := map[string]interface{}{
		"name": "Updated Terminal Name",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/terminal/%s", suite.testTerminalID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.bearerToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code, "Expected 200, got %d: %s", w.Code, w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["status"].(bool))
	assert.Equal(suite.T(), "Updated successfully", response["message"])

	data, ok := response["data"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), suite.testTerminalID, data["id"])
	assert.Equal(suite.T(), "Updated Terminal Name", data["name"])
}

// TestTerminal_ErrorResponses tests that error responses are consistent.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_ErrorResponses() {
	suite.T().Run("Get non-existent terminal by MAC", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/terminal/mac/AA:BB:CC:11:22:33", nil)
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		if status, ok := response["status"]; ok && status != nil {
			assert.False(t, status.(bool))
		}
		if success, ok := response["success"]; ok && success != nil {
			assert.False(t, success.(bool))
		}
		if msg, ok := response["message"]; ok && msg != nil {
			assert.Equal(t, "Terminal not found", msg)
		}
	})

	suite.T().Run("Update non-existent terminal", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Should Fail",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/terminal/non-existent-id", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", suite.bearerToken)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		if status, ok := response["status"]; ok && status != nil {
			assert.False(t, status.(bool))
		}
		if success, ok := response["success"]; ok && success != nil {
			assert.False(t, success.(bool))
		}
		if msg, ok := response["message"]; ok && msg != nil {
			assert.Equal(t, "Not Found", msg)
		}
		assert.NotContains(t, w.Body.String(), "Internal server error")
	})

	suite.T().Run("Get non-existent terminal by ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/terminal/nonexistent-uuid", nil)
		req.Header.Set("Authorization", suite.bearerToken)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404, got %d: %s", w.Code, w.Body.String())
	})
}

// TestTerminal_ValidationErrors tests that validation errors return proper structure.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_ValidationErrors() {
	suite.T().Run("Invalid MAC address format", func(t *testing.T) {
		payload := map[string]string{
			"mac_address":    "invalid-mac",
			"room_id":        "1",
			"name":           "Test Terminal",
			"device_type_id": "1",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected 422, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["status"].(bool))
		assert.Equal(t, "Validation Error", response["message"])
	})

	suite.T().Run("Missing required fields", func(t *testing.T) {
		payload := map[string]string{
			"mac_address": "AA:BB:CC:DD:EE:FF",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected 422, got %d: %s", w.Code, w.Body.String())
	})

	suite.T().Run("Non-numeric room_id", func(t *testing.T) {
		payload := map[string]string{
			"mac_address":    "AA:BB:CC:DD:EE:FF",
			"room_id":        "room-test-001",
			"name":           "Test Terminal",
			"device_type_id": "1",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected 422 (non-numeric room_id), got %d: %s", w.Code, w.Body.String())
	})
}

// TestTerminal_AuthProtectedRoutes tests that protected routes require Bearer token.
func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_AuthProtectedRoutes() {
	suite.T().Run("GET /api/terminal without Bearer returns 401", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/terminal", nil)
		// No Authorization header

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401, got %d: %s", w.Code, w.Body.String())
	})

	suite.T().Run("PUT /api/terminal/:id without Bearer returns 401", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Test",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/api/terminal/some-id", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401, got %d: %s", w.Code, w.Body.String())
	})
}

// TestTerminalBootstrapE2ETest runs the test suite.
func TestTerminalBootstrapE2ETest(t *testing.T) {
	suite.Run(t, new(TerminalBootstrapE2ETestSuite))
}