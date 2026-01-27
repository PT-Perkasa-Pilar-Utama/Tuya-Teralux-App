package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration parameters.
// These are loaded from environment variables or a .env file.
type Config struct {
	// Tuya / general
	TuyaClientID              string
	TuyaClientSecret          string
	TuyaBaseURL               string
	TuyaUserID                string
	ApiKey                    string
	SwaggerBaseURL            string
	GetAllDevicesResponseType string
	CacheTTL                  string

	// Speech / RAG
	OllamaURL        string
	LLMModel         string
	WhisperModelPath string
	MaxFileSize      int64 // bytes
	Port             string

	// Runtime & Networking
	LogLevel   string

	// Database (used for migrations)
	DBType      string
	AutoMigrate string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
}

// AppConfig is the global configuration instance.
var AppConfig *Config

// LoadConfig initializes the AppConfig by loading variables from the environment.
// It searches for a .env file in the current and parent directories if not already set.
// It also triggers an update of the log level based on the loaded configuration.
func LoadConfig() {
	// Load .env first (if present), then overlay with .env.dev if that exists.
	envPath := findEnvFile()
	if envPath == "" {
		log.Println("Warning: .env file not found")
	} else {
		m, err := godotenv.Read(envPath)
		if err != nil {
			log.Println("Warning: Error reading .env file")
		} else {
			for k, v := range m {
				if os.Getenv(k) == "" {
					os.Setenv(k, v)
				}
			}
			log.Printf("Loaded env file: %s", envPath)
		}
	}

	// Attempt to load .env.dev (load values only if not already set in the environment)
	devEnvPath := findFileInParents(".env.dev")
	if devEnvPath != "" {
		m, err := godotenv.Read(devEnvPath)
		if err != nil {
			log.Println("Warning: Error reading .env.dev file")
		} else {
			for k, v := range m {
				if os.Getenv(k) == "" {
					os.Setenv(k, v)
				}
			}
			log.Printf("Loaded env file: %s", devEnvPath)
		}
	}

	// Parse MAX_FILE_SIZE_MB as integer (in MB)
	maxFileSize := int64(0)
	if v := os.Getenv("MAX_FILE_SIZE_MB"); v != "" {
		if mb, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxFileSize = mb * 1024 * 1024
		} else {
			log.Printf("Warning: invalid MAX_FILE_SIZE_MB value '%s'", v)
		}
	}

	AppConfig = &Config{
		TuyaClientID:              os.Getenv("TUYA_CLIENT_ID"),
		TuyaClientSecret:          os.Getenv("TUYA_ACCESS_SECRET"),
		TuyaBaseURL:               os.Getenv("TUYA_BASE_URL"),
		TuyaUserID:                os.Getenv("TUYA_USER_ID"),
		ApiKey:                    os.Getenv("API_KEY"),
		SwaggerBaseURL:            os.Getenv("SWAGGER_BASE_URL"),
		GetAllDevicesResponseType: os.Getenv("GET_ALL_DEVICES_RESPONSE"),
		CacheTTL:                  os.Getenv("CACHE_TTL"),

		OllamaURL:        os.Getenv("OLLAMA_URL"),
		LLMModel:         os.Getenv("LLM_MODEL"),
		WhisperModelPath: os.Getenv("WHISPER_MODEL_PATH"),
		MaxFileSize:      maxFileSize,
		Port:             os.Getenv("PORT"),

		// Runtime
		LogLevel: os.Getenv("LOG_LEVEL"),

		// Database
		DBType:      os.Getenv("DB_TYPE"),
		AutoMigrate: os.Getenv("AUTO_MIGRATE"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
	}

	// Ensure sensible defaults are centralized here
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	UpdateLogLevel()
}

// findEnvFile searches for the .env file in the current directory and up to three parent levels.
//
// return string The path to the .env file if found, otherwise an empty string.
func findEnvFile() string {
	return findFileInParents(".env")
}

// findFileInParents searches for filename in current directory and up to three parent directories.
func findFileInParents(filename string) string {
	path := filename
	if _, err := os.Stat(path); err == nil {
		return path
	}

	for i := 0; i < 3; i++ {
		path = "../" + path
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetConfig returns the singleton AppConfig instance.
// If the config hasn't been loaded, it triggers LoadConfig first.
//
// return *Config The global configuration object.
func GetConfig() *Config {
	if AppConfig == nil {
		LoadConfig()
	}
	return AppConfig
}
