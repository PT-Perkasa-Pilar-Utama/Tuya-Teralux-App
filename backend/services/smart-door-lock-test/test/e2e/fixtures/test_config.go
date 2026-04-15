package fixtures

import (
	"os"
	"time"
)

// TestConfig holds E2E test configuration
type TestConfig struct {
	DeviceID          string
	ClientID          string
	AccessSecret      string
	BaseURL           string
	ValidDurations    []int
	CustomPasswords   []string
	OfflineTestWindow time.Duration
}

// LoadTestConfig loads configuration from environment
func LoadTestConfig() (*TestConfig, error) {
	return &TestConfig{
		DeviceID:          getEnvOrFatal("TUYA_DEVICE_ID"),
		ClientID:          getEnvOrFatal("TUYA_CLIENT_ID"),
		AccessSecret:      getEnvOrFatal("TUYA_ACCESS_SECRET"),
		BaseURL:           getEnv("TUYA_BASE_URL", "https://openapi-sg.iotbing.com"),
		ValidDurations:    []int{5, 60, 1440, 525600}, // 5min, 1hr, 1day, 1year
		CustomPasswords:   []string{"123456", "999999", "000000"},
		OfflineTestWindow: 5 * time.Minute,
	}, nil
}

func getEnvOrFatal(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic("required environment variable " + key + " is not set")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
