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
	TuyaClientID     string
	TuyaClientSecret string
	TuyaBaseURL      string
	TuyaUserID       string
	ApiKey           string
	CacheTTL         string

	// Speech / RAG
	// LLM
	LLMProvider string // "gemini", "orion"

	// Gemini
	GeminiApiKey       string
	GeminiModel        string // Default (legacy/fallback)
	GeminiModelHigh    string // High reasoning
	GeminiModelLow     string // Fast/Low cost
	GeminiModelWhisper string // STT

	// Orion
	OrionBaseURL string
	OrionApiKey  string
	OrionModel   string

	// Whisper
	WhisperLocalModel     string // Path to local model (e.g., bin/ggml-base.bin)
	OrionWhisperBaseURL   string // URL for remote transcription service
	MaxFileSize             int64 // bytes
	Port                    string

	// MQTT
	MqttBroker   string
	MqttUsername string
	MqttPassword string
	MqttTopic    string

	// SMTP
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Runtime & Networking
	LogLevel string

	// Database
	DBType     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}

// AppConfig is the global configuration instance.
var AppConfig *Config

// LoadConfig initializes the AppConfig by loading variables from the environment.
// It searches for a .env file in the current and parent directories if not already set.
// It also triggers an update of the log level based on the loaded configuration.
func LoadConfig() {
	// Load .env (if present)
	envPath := FindEnvFile()
	if envPath == "" {
		log.Println("Warning: .env file not found")
	} else {
		m, err := godotenv.Read(envPath)
		if err != nil {
			log.Println("Warning: Error reading .env file")
		} else {
			for k, v := range m {
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
			log.Printf("Loaded env file: %s", envPath)
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
		TuyaClientID:     os.Getenv("TUYA_CLIENT_ID"),
		TuyaClientSecret: os.Getenv("TUYA_ACCESS_SECRET"),
		TuyaBaseURL:      os.Getenv("TUYA_BASE_URL"),
		TuyaUserID:       os.Getenv("TUYA_USER_ID"),
		ApiKey:    os.Getenv("API_KEY"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		LogLevel:  os.Getenv("LOG_LEVEL"),
		LLMProvider: os.Getenv("LLM_PROVIDER"),

		GeminiApiKey:       os.Getenv("GEMINI_API_KEY"),
		GeminiModelHigh:    os.Getenv("GEMINI_MODEL_HIGH"),
		GeminiModelLow:     os.Getenv("GEMINI_MODEL_LOW"),
		GeminiModelWhisper: os.Getenv("GEMINI_MODEL_WHISPER"),

		OrionBaseURL: os.Getenv("ORION_BASE_URL"),
		OrionApiKey:  os.Getenv("ORION_API_KEY"),
		OrionModel:   os.Getenv("ORION_MODEL"),

		WhisperLocalModel:   os.Getenv("WHISPER_LOCAL_MODEL"),
		OrionWhisperBaseURL: os.Getenv("ORION_WHISPER_BASE_URL"),
		MaxFileSize:             maxFileSize,
		Port:                    os.Getenv("PORT"),

		MqttBroker:   os.Getenv("MQTT_BROKER"),
		MqttUsername: os.Getenv("MQTT_USERNAME"),
		MqttPassword: os.Getenv("MQTT_PASSWORD"),
		MqttTopic:    os.Getenv("MQTT_TOPIC"),

		// SMTP
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),

		// Runtime


		// Database
		DBType:     os.Getenv("DB_TYPE"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}

	// Defaults are removed to enforce explicit configuration via environment variables
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	UpdateLogLevel()
}

// FindEnvFile searches for the .env file in the current directory and up to three parent levels.
//
// return string The path to the .env file if found, otherwise an empty string.
func FindEnvFile() string {
	return FindFileInParents(".env")
}

// FindFileInParents searches for filename in current directory and up to three parent directories.
func FindFileInParents(filename string) string {
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
