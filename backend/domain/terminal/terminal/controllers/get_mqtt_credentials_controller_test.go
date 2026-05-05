package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sensio/domain/common/dtos"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGetOrCreateMQTTCredentialsUseCase mocks the use case for testing
type MockGetOrCreateMQTTCredentialsUseCase struct {
	mock.Mock
}

func (m *MockGetOrCreateMQTTCredentialsUseCase) GetMQTTCredentials(username string) (*services.MQTTCredentials, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.MQTTCredentials), args.Error(1)
}

// GetOrCreateMQTTCredentialsUseCaseInterface defines the interface for the use case
type GetOrCreateMQTTCredentialsUseCaseInterface interface {
	GetMQTTCredentials(username string) (*services.MQTTCredentials, error)
}

// TestableGetMQTTCredentialsController is a testable version that uses an interface
type TestableGetMQTTCredentialsController struct {
	useCase GetOrCreateMQTTCredentialsUseCaseInterface
}

func NewTestableGetMQTTCredentialsController(useCase GetOrCreateMQTTCredentialsUseCaseInterface) *TestableGetMQTTCredentialsController {
	return &TestableGetMQTTCredentialsController{
		useCase: useCase,
	}
}

func (c *TestableGetMQTTCredentialsController) GetMQTTCredentials(ctx *gin.Context) {
	username := ctx.Param("username")

	if username == "" {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
		})
		return
	}

	creds, err := c.useCase.GetMQTTCredentials(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to fetch MQTT credentials",
			Data:    nil,
		})
		return
	}

	if creds == nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "MQTT credentials not found",
			Data:    nil,
		})
		return
	}

	responseDTO := terminal_dtos.MQTTCredentialsResponseDTO{
		Username: creds.Username,
		Password: creds.Password,
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Credentials retrieved successfully",
		Data:    responseDTO,
	})
}

func TestGetMQTTCredentials_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := new(MockGetOrCreateMQTTCredentialsUseCase)
	controller := NewTestableGetMQTTCredentialsController(mockUseCase)

	username := "device_001"
	mockCreds := &services.MQTTCredentials{
		Username: username,
		Password: "secret_password_123",
	}

	mockUseCase.On("GetMQTTCredentials", username).Return(mockCreds, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/mqtt/users/"+username, nil)
	c.Params = gin.Params{{Key: "username", Value: username}}

	controller.GetMQTTCredentials(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dtos.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Status)
	assert.Equal(t, "Credentials retrieved successfully", resp.Message)

	mockUseCase.AssertExpectations(t)
}

func TestGetMQTTCredentials_TerminalNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := new(MockGetOrCreateMQTTCredentialsUseCase)
	controller := NewTestableGetMQTTCredentialsController(mockUseCase)

	username := "nonexistent_device"

	mockUseCase.On("GetMQTTCredentials", username).Return(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/mqtt/users/"+username, nil)
	c.Params = gin.Params{{Key: "username", Value: username}}

	controller.GetMQTTCredentials(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp dtos.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "MQTT credentials not found", resp.Message)

	mockUseCase.AssertExpectations(t)
}

func TestGetMQTTCredentials_ServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := new(MockGetOrCreateMQTTCredentialsUseCase)
	controller := NewTestableGetMQTTCredentialsController(mockUseCase)

	username := "device_error"
	mockError := errors.New("rust auth service unavailable")

	mockUseCase.On("GetMQTTCredentials", username).Return(nil, mockError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/mqtt/users/"+username, nil)
	c.Params = gin.Params{{Key: "username", Value: username}}

	controller.GetMQTTCredentials(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp dtos.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Failed to fetch MQTT credentials", resp.Message)

	mockUseCase.AssertExpectations(t)
}

func TestGetMQTTCredentials_ValidationError_EmptyUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := new(MockGetOrCreateMQTTCredentialsUseCase)
	controller := NewTestableGetMQTTCredentialsController(mockUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/mqtt/users/", nil)
	c.Params = gin.Params{{Key: "username", Value: ""}}

	controller.GetMQTTCredentials(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dtos.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Validation Error", resp.Message)
}
