package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"sensio/domain/common/middlewares"
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/terminal/terminal/controllers"
	"sensio/domain/terminal/terminal/entities"
	"sensio/domain/terminal/terminal/repositories"
	"sensio/domain/terminal/terminal/services"
	terminal_usecases "sensio/domain/terminal/terminal/usecases"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TerminalBootstrapE2ETestSuite struct {
	suite.Suite
	router             *gin.Engine
	db                 *gorm.DB
	apiKey             string
	validBearerToken   string
	testTerminalID     string
	badger             *infrastructure.BadgerService
	mockEmqxServer     *httptest.Server
	mockEmqxCreds      map[string]string
	mockEmqxCredsMutex sync.Mutex
}

type mockTuyaTokenProvider struct{}

func (m *mockTuyaTokenProvider) GetTuyaAccessToken() (string, error) {
	return "mock-tuya-token", nil
}

func (suite *TerminalBootstrapE2ETestSuite) SetupSuite() {
	_ = os.Setenv("GO_TEST", "true")
	_ = os.Setenv("API_KEY", "test-api-key")
	_ = os.Setenv("JWT_SECRET", "test-jwt-secret-for-e2e")
	_ = os.Setenv("EMQX_AUTH_API_KEY", "mock-emqx-key")

	utils.AppConfig = nil

	suite.mockEmqxCreds = make(map[string]string)
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.mockEmqxCredsMutex.Lock()
		defer suite.mockEmqxCredsMutex.Unlock()

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			if r.URL.Path == "/mqtt/create" {
				var req struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"success":false,"message":"bad request"}`, http.StatusBadRequest)
					return
				}
				if _, exists := suite.mockEmqxCreds[req.Username]; exists {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"message": "user already exists",
					})
					return
				}
				suite.mockEmqxCreds[req.Username] = req.Password
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"message": "user created",
					"data": map[string]string{
						"username": req.Username,
						"password": req.Password,
					},
				})
				return
			}

		case http.MethodGet:
			if len(r.URL.Path) > 14 && r.URL.Path[:14] == "/mqtt/users/" {
				username := r.URL.Path[14:]
				if password, exists := suite.mockEmqxCreds[username]; exists {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": true,
						"data": map[string]string{
							"username": username,
							"password": password,
						},
					})
					return
				}
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "user not found",
				})
				return
			}
		}
		http.NotFound(w, r)
	})

	suite.mockEmqxServer = httptest.NewServer(mockHandler)

	_ = os.Setenv("EMQX_AUTH_BASE_URL", suite.mockEmqxServer.URL)
	utils.LoadConfig()

	testDB, err := infrastructure.InitDB()
	if err != nil {
		suite.T().Fatalf("Failed to initialize database: %v", err)
	}
	suite.db = testDB
	infrastructure.DB = testDB

	badger, err := infrastructure.NewBadgerService("./tmp/badger_e2e_test")
	if err != nil {
		suite.T().Fatalf("Failed to initialize BadgerDB: %v", err)
	}
	suite.badger = badger

	err = testDB.AutoMigrate(&entities.Terminal{}, &entities.MQTTUser{})
	if err != nil {
		suite.T().Fatalf("Failed to migrate test tables: %v", err)
	}

	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	suite.registerRoutes()

	suite.validBearerToken, err = utils.GenerateToken("test-terminal-user")
	if err != nil {
		suite.T().Fatalf("Failed to generate test Bearer token: %v", err)
	}

	suite.apiKey = "test-api-key"
}

func (suite *TerminalBootstrapE2ETestSuite) TearDownSuite() {
	if suite.testTerminalID != "" {
		suite.db.Delete(&entities.Terminal{}, "id = ?", suite.testTerminalID)
	}

	if suite.badger != nil {
		_ = suite.badger.Close()
	}

	if suite.mockEmqxServer != nil {
		suite.mockEmqxServer.Close()
	}

	sqlDB, err := suite.db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	_ = os.Unsetenv("GO_TEST")
	_ = os.Unsetenv("API_KEY")
	_ = os.Unsetenv("JWT_SECRET")
	_ = os.Unsetenv("EMQX_AUTH_BASE_URL")
	_ = os.Unsetenv("EMQX_AUTH_API_KEY")
}

func (suite *TerminalBootstrapE2ETestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// Cleanup deterministic fixtures to make reruns idempotent
	suite.db.Exec("DELETE FROM terminal WHERE mac_address IN (?, ?)",
		"AA:BB:CC:DD:EE:FF", "CC:DD:EE:FF:AA:BB")
	suite.db.Exec("DELETE FROM mqtt_users WHERE username IN (?, ?)",
		"AA:BB:CC:DD:EE:FF", "CC:DD:EE:FF:AA:BB")

	if suite.testTerminalID != "" {
		suite.db.Delete(&entities.Terminal{}, "id = ?", suite.testTerminalID)
		suite.testTerminalID = ""
	}

	suite.mockEmqxCredsMutex.Lock()
	suite.mockEmqxCreds = make(map[string]string)
	suite.mockEmqxCredsMutex.Unlock()
}

func (suite *TerminalBootstrapE2ETestSuite) registerRoutes() {
	publicGroup := suite.router.Group("/")
	publicGroup.Use(middlewares.ApiKeyMiddleware())

	protected := suite.router.Group("/")
	protected.Use(middlewares.AuthMiddleware(&mockTuyaTokenProvider{}))

	terminalRepo := repositories.NewTerminalRepository(suite.badger)

	macExternalSvc := services.NewMacRegistrationExternalService()
	mqttClient := services.NewMqttAuthClient(
		utils.GetConfig().EmqxAuthBaseURL,
		utils.GetConfig().EmqxAuthApiKey,
	)

	createUC := terminal_usecases.NewCreateTerminalUseCase(terminalRepo, macExternalSvc, mqttClient)
	getByMacUC := terminal_usecases.NewGetTerminalByMACUseCase(terminalRepo, mqttClient)
	updateUC := terminal_usecases.NewUpdateTerminalUseCase(terminalRepo)
	getMqttCtrl := controllers.NewGetMQTTCredentialsController(mqttClient)

	createCtrl := controllers.NewCreateTerminalController(createUC)
	getByMacCtrl := controllers.NewGetTerminalByMACController(getByMacUC)
	updateCtrl := controllers.NewUpdateTerminalController(updateUC)

	publicGroup.POST("/api/terminal", createCtrl.CreateTerminal)
	publicGroup.GET("/api/terminal/mac/:mac", getByMacCtrl.GetTerminalByMAC)

	protected.PUT("/api/terminal/:id", updateCtrl.UpdateTerminal)
	protected.GET("/api/mqtt/users/:username", getMqttCtrl.GetMQTTCredentials)
}

func (suite *TerminalBootstrapE2ETestSuite) TestTerminalBootstrap_FullFlow() {
	// Skip - depends on external API aplikasi-big.com which is unreachable in test env
	suite.T().Skip("Skipped: requires external API 'aplikasi-big.com' which is unreachable in test environment")
}

func (suite *TerminalBootstrapE2ETestSuite) TestTerminal_ValidationErrors() {
	suite.T().Run("Invalid MAC address format", func(t *testing.T) {
		payload := map[string]interface{}{
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

		require.Equal(t, http.StatusUnprocessableEntity, w.Code,
			"Expected 422 for invalid MAC, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.False(t, response["status"].(bool),
			"Expected status=false for validation error")
		require.Equal(t, "Validation Error", response["message"],
			"Expected 'Validation Error' message")
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

		require.Equal(t, http.StatusUnprocessableEntity, w.Code,
			"Expected 422 for missing required fields, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.False(t, response["status"].(bool),
			"Expected status=false for validation error")
		require.Equal(t, "Validation Error", response["message"],
			"Expected 'Validation Error' message")
	})

	suite.T().Run("Non-numeric room_id triggers validation", func(t *testing.T) {
		payload := map[string]interface{}{
			"mac_address":    "CC:DD:EE:FF:AA:BB",
			"room_id":        "not-a-number",
			"name":           "Test Terminal",
			"device_type_id": "1",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/terminal", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", suite.apiKey)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code,
			"Expected 422 for non-numeric room_id, got %d: %s", w.Code, w.Body.String())

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.False(t, response["status"].(bool),
			"Expected status=false for validation error")
	})
}

func TestTerminalBootstrapE2ETest(t *testing.T) {
	suite.Run(t, new(TerminalBootstrapE2ETestSuite))
}
