package domain

import (
	"testing"
	"time"
)

func TestPassword_IsValid_NotExpired(t *testing.T) {
	// Arrange
	password := &Password{
		Value:        "123456",
		Type:         PasswordTypeDynamic,
		ExpireAt:     time.Now().Add(5 * time.Minute),
		ValidMinutes: 5,
	}

	// Act
	valid := password.IsValid()

	// Assert
	if !valid {
		t.Error("expected password to be valid, got false")
	}
}

func TestPassword_IsValid_Expired(t *testing.T) {
	// Arrange
	password := &Password{
		Value:        "123456",
		Type:         PasswordTypeDynamic,
		ExpireAt:     time.Now().Add(-5 * time.Minute),
		ValidMinutes: 5,
	}

	// Act
	valid := password.IsValid()

	// Assert
	if valid {
		t.Error("expected password to be invalid (expired), got true")
	}
}

func TestPassword_IsValid_ExactlyAtExpiration(t *testing.T) {
	// Arrange
	password := &Password{
		Value:        "123456",
		Type:         PasswordTypeDynamic,
		ExpireAt:     time.Now(),
		ValidMinutes: 5,
	}

	// Act
	valid := password.IsValid()

	// Assert
	// time.Now().Before() returns false for equal times, so expired
	if valid {
		t.Error("expected password at expiration boundary to be invalid, got true")
	}
}

func TestPassword_TimeRemaining_NotExpired(t *testing.T) {
	// Arrange
	futureTime := time.Now().Add(10 * time.Minute)
	password := &Password{
		Value:        "123456",
		Type:         PasswordTypeDynamic,
		ExpireAt:     futureTime,
		ValidMinutes: 10,
	}

	// Act
	remaining := password.TimeRemaining()

	// Assert
	if remaining <= 0 {
		t.Errorf("expected positive time remaining, got %v", remaining)
	}

	// Should be approximately 10 minutes (allow 1 second tolerance for test execution time)
	if remaining > 10*time.Minute+time.Second {
		t.Errorf("expected time remaining ~10 minutes, got %v", remaining)
	}
}

func TestPassword_TimeRemaining_Expired(t *testing.T) {
	// Arrange
	pastTime := time.Now().Add(-5 * time.Minute)
	password := &Password{
		Value:        "123456",
		Type:         PasswordTypeDynamic,
		ExpireAt:     pastTime,
		ValidMinutes: 5,
	}

	// Act
	remaining := password.TimeRemaining()

	// Assert
	if remaining > 0 {
		t.Errorf("expected negative time remaining for expired password, got %v", remaining)
	}
}

func TestPasswordType_Constants(t *testing.T) {
	// Test PasswordTypeDynamic
	if PasswordTypeDynamic != "dynamic" {
		t.Errorf("expected PasswordTypeDynamic='dynamic', got '%s'", PasswordTypeDynamic)
	}

	// Test PasswordTypeTemporary
	if PasswordTypeTemporary != "temporary" {
		t.Errorf("expected PasswordTypeTemporary='temporary', got '%s'", PasswordTypeTemporary)
	}
}

func TestPasswordRequest_Validate_ValidDynamic(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type:     PasswordTypeDynamic,
		DeviceID: "test-device-123",
	}

	// Act
	err := req.Validate()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPasswordRequest_Validate_ValidTemporary(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type:     PasswordTypeTemporary,
		DeviceID: "test-device-123",
		Duration: 60,
	}

	// Act
	err := req.Validate()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPasswordRequest_Validate_MissingDeviceID(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type: PasswordTypeDynamic,
	}

	// Act
	err := req.Validate()

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrDeviceIDRequired {
		t.Errorf("expected ErrDeviceIDRequired, got %v", err)
	}
}

func TestPasswordRequest_Validate_EmptyDeviceID(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type:     PasswordTypeDynamic,
		DeviceID: "",
	}

	// Act
	err := req.Validate()

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrDeviceIDRequired {
		t.Errorf("expected ErrDeviceIDRequired, got %v", err)
	}
}

func TestPasswordRequest_Validate_TemporaryZeroDuration(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type:     PasswordTypeTemporary,
		DeviceID: "test-device-123",
		Duration: 0,
	}

	// Act
	err := req.Validate()

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrInvalidDuration {
		t.Errorf("expected ErrInvalidDuration, got %v", err)
	}
}

func TestPasswordRequest_Validate_TemporaryNegativeDuration(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type:     PasswordTypeTemporary,
		DeviceID: "test-device-123",
		Duration: -10,
	}

	// Act
	err := req.Validate()

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrInvalidDuration {
		t.Errorf("expected ErrInvalidDuration, got %v", err)
	}
}

func TestPasswordRequest_Validate_TemporaryWithCustomPassword(t *testing.T) {
	// Arrange
	req := &PasswordRequest{
		Type:        PasswordTypeTemporary,
		DeviceID:    "test-device-123",
		Duration:    60,
		CustomValue: "123456",
	}

	// Act
	err := req.Validate()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidationError_Error(t *testing.T) {
	// Arrange
	err := &ValidationError{Message: "test error message"}

	// Act
	msg := err.Error()

	// Assert
	if msg != "test error message" {
		t.Errorf("expected 'test error message', got '%s'", msg)
	}
}

func TestErrorConstants(t *testing.T) {
	// Test ErrDeviceIDRequired
	if ErrDeviceIDRequired == nil {
		t.Error("expected ErrDeviceIDRequired to be defined")
	}
	if ErrDeviceIDRequired.Error() != "device_id is required" {
		t.Errorf("expected 'device_id is required', got '%s'", ErrDeviceIDRequired.Error())
	}

	// Test ErrInvalidDuration
	if ErrInvalidDuration == nil {
		t.Error("expected ErrInvalidDuration to be defined")
	}
	if ErrInvalidDuration.Error() != "duration must be positive" {
		t.Errorf("expected 'duration must be positive', got '%s'", ErrInvalidDuration.Error())
	}

	// Test ErrPasswordExpired
	if ErrPasswordExpired == nil {
		t.Error("expected ErrPasswordExpired to be defined")
	}
	if ErrPasswordExpired.Error() != "password has expired" {
		t.Errorf("expected 'password has expired', got '%s'", ErrPasswordExpired.Error())
	}
}
