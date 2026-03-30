package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Tuya TuyaConfig
}

// TuyaConfig holds Tuya-specific configuration
type TuyaConfig struct {
	ClientID     string
	AccessSecret string
	BaseURL      string
	UserID       string
	DeviceID     string
}

// Load loads configuration from environment and .env file
func Load() (*Config, error) {
	// Load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Tuya: TuyaConfig{
			ClientID:     getEnv("TUYA_CLIENT_ID"),
			AccessSecret: getEnv("TUYA_ACCESS_SECRET"),
			BaseURL:      getEnv("TUYA_BASE_URL", "https://openapi-sg.iotbing.com"),
			UserID:       getEnv("TUYA_USER_ID"),
			DeviceID:     getEnv("TUYA_DEVICE_ID"),
		},
	}

	// Set defaults if empty
	if cfg.Tuya.BaseURL == "" {
		cfg.Tuya.BaseURL = "https://openapi-sg.iotbing.com"
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Tuya.ClientID == "" {
		return &ConfigError{"TUYA_CLIENT_ID is required"}
	}

	if c.Tuya.AccessSecret == "" {
		return &ConfigError{"TUYA_ACCESS_SECRET is required"}
	}

	if c.Tuya.UserID == "" {
		return &ConfigError{"TUYA_USER_ID is required"}
	}

	if c.Tuya.DeviceID == "" {
		return &ConfigError{"TUYA_DEVICE_ID is required"}
	}

	return nil
}

// getEnv gets environment variable with optional default
func getEnv(key string, defaultValue ...string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
