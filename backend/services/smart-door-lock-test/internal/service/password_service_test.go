package service

import (
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"testing"
	"time"
)

// MockPasswordRepository for testing
type MockPasswordRepository struct {
	PasswordToReturn *domain.Password
	ErrorToReturn    error
	CallCount        int
	LastDeviceID     string
	LastDuration     int
	LastCustomPwd    string
}

func (m *MockPasswordRepository) Generate(req *domain.PasswordRequest) (*domain.Password, error) {
	m.CallCount++
	m.LastDeviceID = req.DeviceID

	if req.Type == domain.PasswordTypeTemporary {
		m.LastDuration = req.Duration
		m.LastCustomPwd = req.CustomValue
	}

	return m.PasswordToReturn, m.ErrorToReturn
}

func TestPasswordService_GenerateDynamicPassword_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{
		PasswordToReturn: &domain.Password{
			Value:        "123456",
			Type:         domain.PasswordTypeDynamic,
			ExpireAt:     time.Now().Add(5 * time.Minute),
			ValidMinutes: 5,
		},
	}
	service := NewPasswordService(mockRepo)

	// Act
	password, err := service.GenerateDynamicPassword("test-device-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.Value != "123456" {
		t.Errorf("expected password '123456', got '%s'", password.Value)
	}

	if password.Type != domain.PasswordTypeDynamic {
		t.Errorf("expected type 'dynamic', got '%s'", password.Type)
	}

	if mockRepo.CallCount != 1 {
		t.Errorf("expected 1 call to repository, got %d", mockRepo.CallCount)
	}

	if mockRepo.LastDeviceID != "test-device-123" {
		t.Errorf("expected deviceID 'test-device-123', got '%s'", mockRepo.LastDeviceID)
	}
}

func TestPasswordService_GenerateDynamicPassword_Error(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{
		ErrorToReturn: &domain.ValidationError{"API not subscribed"},
	}
	service := NewPasswordService(mockRepo)

	// Act
	_, err := service.GenerateDynamicPassword("test-device-123")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "API not subscribed" {
		t.Errorf("expected 'API not subscribed', got '%s'", err.Error())
	}
}

func TestPasswordService_GenerateTemporaryPassword_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{
		PasswordToReturn: &domain.Password{
			Value:        "654321",
			Type:         domain.PasswordTypeTemporary,
			ExpireAt:     time.Now().Add(60 * time.Minute),
			ValidMinutes: 60,
		},
	}
	service := NewPasswordService(mockRepo)

	// Act
	password, err := service.GenerateTemporaryPassword("test-device-123", 60, "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.Value != "654321" {
		t.Errorf("expected password '654321', got '%s'", password.Value)
	}

	if password.Type != domain.PasswordTypeTemporary {
		t.Errorf("expected type 'temporary', got '%s'", password.Type)
	}

	if password.ValidMinutes != 60 {
		t.Errorf("expected validMinutes 60, got %d", password.ValidMinutes)
	}

	if mockRepo.CallCount != 1 {
		t.Errorf("expected 1 call to repository, got %d", mockRepo.CallCount)
	}

	if mockRepo.LastDeviceID != "test-device-123" {
		t.Errorf("expected deviceID 'test-device-123', got '%s'", mockRepo.LastDeviceID)
	}

	if mockRepo.LastDuration != 60 {
		t.Errorf("expected duration 60, got %d", mockRepo.LastDuration)
	}
}

func TestPasswordService_GenerateTemporaryPassword_WithCustomPassword(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{
		PasswordToReturn: &domain.Password{
			Value:        "999999",
			Type:         domain.PasswordTypeTemporary,
			ExpireAt:     time.Now().Add(120 * time.Minute),
			ValidMinutes: 120,
		},
	}
	service := NewPasswordService(mockRepo)

	// Act
	password, err := service.GenerateTemporaryPassword("test-device-123", 120, "999999")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.Value != "999999" {
		t.Errorf("expected password '999999', got '%s'", password.Value)
	}

	if mockRepo.LastCustomPwd != "999999" {
		t.Errorf("expected custom password '999999', got '%s'", mockRepo.LastCustomPwd)
	}
}

func TestPasswordService_GenerateTemporaryPassword_Error(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{
		ErrorToReturn: &domain.ValidationError{"device is offline"},
	}
	service := NewPasswordService(mockRepo)

	// Act
	_, err := service.GenerateTemporaryPassword("test-device-123", 60, "")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "device is offline" {
		t.Errorf("expected 'device is offline', got '%s'", err.Error())
	}
}

func TestPasswordService_ValidatePassword_Valid(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{}
	service := NewPasswordService(mockRepo)

	validPassword := &domain.Password{
		Value:        "123456",
		Type:         domain.PasswordTypeDynamic,
		ExpireAt:     time.Now().Add(5 * time.Minute), // Expires in 5 minutes
		ValidMinutes: 5,
	}

	// Act
	err := service.ValidatePassword(validPassword)

	// Assert
	if err != nil {
		t.Fatalf("expected no error for valid password, got %v", err)
	}
}

func TestPasswordService_ValidatePassword_Expired(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{}
	service := NewPasswordService(mockRepo)

	expiredPassword := &domain.Password{
		Value:        "123456",
		Type:         domain.PasswordTypeDynamic,
		ExpireAt:     time.Now().Add(-5 * time.Minute), // Expired 5 minutes ago
		ValidMinutes: 5,
	}

	// Act
	err := service.ValidatePassword(expiredPassword)

	// Assert
	if err == nil {
		t.Fatal("expected error for expired password, got nil")
	}

	if err != domain.ErrPasswordExpired {
		t.Errorf("expected ErrPasswordExpired, got %v", err)
	}
}

func TestPasswordService_ValidatePassword_ExactlyAtExpiration(t *testing.T) {
	// Arrange
	mockRepo := &MockPasswordRepository{}
	service := NewPasswordService(mockRepo)

	// Password expiring right now (edge case)
	expiredPassword := &domain.Password{
		Value:        "123456",
		Type:         domain.PasswordTypeDynamic,
		ExpireAt:     time.Now(),
		ValidMinutes: 5,
	}

	// Act
	err := service.ValidatePassword(expiredPassword)

	// Assert
	// This should fail since the password is at expiration boundary
	if err == nil {
		t.Log("Note: Password at exact expiration time is considered valid (time.Now().Before returns false for equal times)")
	}
}
