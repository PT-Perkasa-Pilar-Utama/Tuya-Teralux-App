package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration parameters.
// These are loaded from environment variables or a .env file.
type Config struct {
	// Tuya / general
	TuyaClientID           string
	TuyaClientSecret       string
	TuyaBaseURL            string
	TuyaUserID             string
	ApiKey                 string
	CacheTTL               string
	ApplicationEnvironment string

	// Speech / RAG
	// LLM
	LLMProvider string // "gemini", "orion"

	// Gemini
	GeminiApiKey       string
	GeminiModelHigh    string // High reasoning
	GeminiModelLow     string // Fast/Low cost
	GeminiModelWhisper string // STT

	// OpenAI
	OpenAIApiKey       string
	OpenAIModelHigh    string
	OpenAIModelLow     string
	OpenAIModelWhisper string

	// Groq
	GroqApiKey       string
	GroqModelHigh    string
	GroqModelLow     string
	GroqModelWhisper string

	// Orion
	OrionBaseURL string
	OrionApiKey  string
	OrionModel   string

	// BIG API
	BIGAPIBaseURL string // BIGAPIBaseURL is the base URL for the BIG API (aplikasi-big.com)

	// Local Models
	WhisperLocalModel   string // Path to whisper ggml model
	LlamaLocalModel     string // Path to llama gguf model (e.g., bin/ggml-base.bin)
	OrionWhisperBaseURL string // URL for remote transcription service
	MaxFileSize         int64  // bytes
	Port                string

	// Python Services (for models-v1 proxy)
	PythonWhisperServiceURL  string // URL for Python Whisper service (HTTP - legacy)
	PythonRAGServiceURL      string // URL for Python RAG service (HTTP - legacy)
	PythonPipelineServiceURL string // URL for Python Pipeline service (HTTP - legacy)
	PythonGrpcServiceURL     string // URL for Python gRPC service (e.g., "localhost:50051")

	// MQTT
	MqttBroker      string
	MqttUsername    string
	MqttPassword    string
	EmqxAuthBaseURL string
	EmqxAuthApiKey  string

	// SMTP
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// WhatsApp Notification
	WANotificationBaseURL string

	// Runtime & Networking
	LogLevel string

	// Storage (S3)
	S3Enabled             bool
	S3Bucket              string
	S3Region              string
	S3Prefix              string
	S3SignedURLTTLSeconds int
	AWSAccessKeyID        string
	AWSSecretAccessKey    string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string

	// Chunk Upload & Async Tasks
	EnableChunkUpload          bool
	EnableSignedUpload         bool
	ChunkUploadDefaultChunkMB  int
	ChunkUploadMinChunkMB      int
	ChunkUploadMaxChunkMB      int
	ChunkUploadMaxFileSizeGB   int
	ChunkUploadSessionTTL      string
	ChunkUploadCleanupInterval string
	TranscribeAsyncTimeout     string
	PipelineAsyncTimeout       string
	TaskStatusTTL              string

	// Byte-based chunk upload config (new, takes precedence over MB-based)
	ChunkUploadDefaultChunkBytes int64
	ChunkUploadMinChunkBytes     int64
	ChunkUploadMaxChunkBytes     int64

	// Audio Segmentation & Pipeline Config
	AudioSegmentEnabled        bool
	AudioSegmentSec            int
	AudioSegmentOverlapSec     int
	AudioSegmentMaxConcurrency int
	TaskEventPublishEnabled    bool
	OrionTranscribeTimeout     string
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
			isTest := os.Getenv("GO_TEST") == "true"
			for k, v := range m {
				// APPLICATION_ENVIRONMENT: always preserve if already set in environment (e.g., via docker-compose)
				if k == "APPLICATION_ENVIRONMENT" && os.Getenv("APPLICATION_ENVIRONMENT") != "" {
					continue
				}
				// In production/dev, we overwrite OS vars with .env values (previous behavior)
				// In tests, we preserve what the test setup (e.g. TestApiKeyMiddleware) has configured
				if !isTest {
					if err := os.Setenv(k, v); err != nil {
						log.Printf("Warning: failed to set environment variable %s", k)
					}
				} else {
					// Test mode: only set if not present
					if _, exists := os.LookupEnv(k); !exists {
						if err := os.Setenv(k, v); err != nil {
							log.Printf("Warning: failed to set environment variable %s", k)
						}
					}
				}
			}
			if isTest {
				log.Printf("Loaded env file in TEST mode (preserving existing variables): %s", envPath)
			} else {
				log.Printf("Loaded env file and overwrote environment variables: %s", envPath)
			}
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
		TuyaClientID:           os.Getenv("TUYA_CLIENT_ID"),
		TuyaClientSecret:       os.Getenv("TUYA_ACCESS_SECRET"),
		TuyaBaseURL:            os.Getenv("TUYA_BASE_URL"),
		TuyaUserID:             os.Getenv("TUYA_USER_ID"),
		ApiKey:                 os.Getenv("API_KEY"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		LogLevel:               os.Getenv("LOG_LEVEL"),
		ApplicationEnvironment: os.Getenv("APPLICATION_ENVIRONMENT"),
		S3Enabled:              getEnvAsBool("S3_ENABLED", false),
		S3Bucket:               os.Getenv("S3_BUCKET"),
		S3Region:               os.Getenv("S3_REGION"),
		S3Prefix:               getEnvAsDefault("S3_PREFIX", "Sensio/"),
		S3SignedURLTTLSeconds:  getEnvAsInt("S3_SIGNED_URL_TTL_SECONDS", 300),
		AWSAccessKeyID:         os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:     os.Getenv("AWS_SECRET_ACCESS_KEY"),
		LLMProvider:            os.Getenv("LLM_PROVIDER"),

		GeminiApiKey:       os.Getenv("GEMINI_API_KEY"),
		GeminiModelHigh:    os.Getenv("GEMINI_MODEL_HIGH"),
		GeminiModelLow:     os.Getenv("GEMINI_MODEL_LOW"),
		GeminiModelWhisper: os.Getenv("GEMINI_MODEL_WHISPER"),

		OpenAIApiKey:       os.Getenv("OPENAI_API_KEY"),
		OpenAIModelHigh:    os.Getenv("OPENAI_MODEL_HIGH"),
		OpenAIModelLow:     os.Getenv("OPENAI_MODEL_LOW"),
		OpenAIModelWhisper: os.Getenv("OPENAI_MODEL_WHISPER"),

		GroqApiKey:       os.Getenv("GROQ_API_KEY"),
		GroqModelHigh:    os.Getenv("GROQ_MODEL_HIGH"),
		GroqModelLow:     os.Getenv("GROQ_MODEL_LOW"),
		GroqModelWhisper: os.Getenv("GROQ_MODEL_WHISPER"),

		OrionBaseURL: os.Getenv("ORION_BASE_URL"),
		OrionApiKey:  os.Getenv("ORION_API_KEY"),
		OrionModel:   os.Getenv("ORION_MODEL"),

		// BIG API
		BIGAPIBaseURL: getEnvAsDefault("BIG_API_BASE_URL", "https://aplikasi-big.com"),

		// Local Models
		WhisperLocalModel:   os.Getenv("WHISPER_LOCAL_MODEL"),
		LlamaLocalModel:     os.Getenv("LLAMA_LOCAL_MODEL"),
		OrionWhisperBaseURL: os.Getenv("ORION_WHISPER_BASE_URL"),
		MaxFileSize:         maxFileSize,
		Port:                os.Getenv("PORT"),

		// Python Services (for models-v1 proxy)
		PythonWhisperServiceURL:  getEnvAsDefault("PYTHON_WHISPER_SERVICE_URL", "http://localhost:8001"),
		PythonRAGServiceURL:      getEnvAsDefault("PYTHON_RAG_SERVICE_URL", "http://localhost:8001"),
		PythonPipelineServiceURL: getEnvAsDefault("PYTHON_PIPELINE_SERVICE_URL", "http://localhost:8001"),
		PythonGrpcServiceURL:     getEnvAsDefault("PYTHON_GRPC_SERVICE_URL", "localhost:50051"),

		MqttBroker:      os.Getenv("MQTT_BROKER"),
		MqttUsername:    os.Getenv("MQTT_USERNAME"),
		MqttPassword:    os.Getenv("MQTT_PASSWORD"),
		EmqxAuthBaseURL: os.Getenv("EMQX_AUTH_BASE_URL"),
		EmqxAuthApiKey:  os.Getenv("EMQX_AUTH_API_KEY"),

		// SMTP
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),

		// WhatsApp Notification
		WANotificationBaseURL: getEnvAsDefault("WA_NOTIFICATION_BASE_URL", "http://10.10.3.24:3000/api/v1/send"),

		// Runtime

		// Database
		DBHost:     os.Getenv("MYSQL_HOST"),
		DBPort:     os.Getenv("MYSQL_PORT"),
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBName:     os.Getenv("MYSQL_DATABASE"),

		// Chunk Upload & Async Tasks
		EnableChunkUpload:          getEnvAsBool("ENABLE_CHUNK_UPLOAD", false),
		EnableSignedUpload:         getEnvAsBool("ENABLE_SIGNED_UPLOAD", false),
		ChunkUploadDefaultChunkMB:  getEnvAsInt("CHUNK_UPLOAD_DEFAULT_CHUNK_MB", 8),
		ChunkUploadMinChunkMB:      getEnvAsInt("CHUNK_UPLOAD_MIN_CHUNK_MB", 1),
		ChunkUploadMaxChunkMB:      getEnvAsInt("CHUNK_UPLOAD_MAX_CHUNK_MB", 32),
		ChunkUploadMaxFileSizeGB:   getEnvAsInt("CHUNK_UPLOAD_MAX_FILE_SIZE_GB", 20),
		ChunkUploadSessionTTL:      getEnvDuration("CHUNK_UPLOAD_SESSION_TTL"),
		ChunkUploadCleanupInterval: getEnvDuration("CHUNK_UPLOAD_CLEANUP_INTERVAL"),
		TranscribeAsyncTimeout:     getEnvDuration("TRANSCRIBE_ASYNC_TIMEOUT"),
		PipelineAsyncTimeout:       getEnvDuration("PIPELINE_ASYNC_TIMEOUT"),
		TaskStatusTTL:              getEnvDuration("TASK_STATUS_TTL"),

		// Byte-based chunk upload config (new, takes precedence over MB-based)
		// Resolution order: 1) byte-based env var, 2) legacy MB env var, 3) hardcoded default
		ChunkUploadDefaultChunkBytes: getByteConfigWithMBFallback(
			"CHUNK_UPLOAD_DEFAULT_CHUNK_BYTES",
			"CHUNK_UPLOAD_DEFAULT_CHUNK_MB",
			1*1024*1024, // default: 1 MB
		),
		ChunkUploadMinChunkBytes: getByteConfigWithMBFallback(
			"CHUNK_UPLOAD_MIN_CHUNK_BYTES",
			"CHUNK_UPLOAD_MIN_CHUNK_MB",
			256*1024, // default: 256 KB (lowered from 1 MB)
		),
		ChunkUploadMaxChunkBytes: getByteConfigWithMBFallback(
			"CHUNK_UPLOAD_MAX_CHUNK_BYTES",
			"CHUNK_UPLOAD_MAX_CHUNK_MB",
			32*1024*1024, // default: 32 MB
		),

		// Audio Segmentation & Pipeline
		AudioSegmentEnabled:        os.Getenv("AUDIO_SEGMENT_ENABLED") == "true",
		AudioSegmentSec:            getEnvAsInt("AUDIO_SEGMENT_SEC", 600),
		AudioSegmentOverlapSec:     getEnvAsInt("AUDIO_SEGMENT_OVERLAP_SEC", 2),
		AudioSegmentMaxConcurrency: getEnvAsInt("AUDIO_SEGMENT_MAX_CONCURRENCY", 2),
		TaskEventPublishEnabled:    os.Getenv("TASK_EVENT_PUBLISH_ENABLED") == "true",
		OrionTranscribeTimeout:     getEnvAsDefault("ORION_TRANSCRIBE_TIMEOUT", "360s"),
	}

	// Defaults are removed to enforce explicit configuration via environment variables
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	AppConfig.S3Prefix = normalizeS3Prefix(AppConfig.S3Prefix)
	validateS3RequiredEnv(AppConfig)

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

// getEnvAsInt reads an environment variable and returns its integer value or a default.
func getEnvAsInt(key string, defaultVal int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultVal
}

// getEnvAsBool reads an environment variable and returns its boolean value or a default.
func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultVal
	}

	return value
}

// getByteConfigWithMBFallback reads byte-based config with MB fallback.
// Resolution order: 1) byte-based env var, 2) legacy MB env var, 3) hardcoded default.
func getByteConfigWithMBFallback(byteKey string, mbKey string, defaultBytes int64) int64 {
	// 1. Try byte-based env var first
	if valueStr := os.Getenv(byteKey); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			return value
		}
	}

	// 2. Try legacy MB env var
	if valueStr := os.Getenv(mbKey); valueStr != "" {
		if mb, err := strconv.Atoi(valueStr); err == nil {
			return int64(mb) * 1024 * 1024
		}
	}

	// 3. Return hardcoded default
	return defaultBytes
}

// getEnvAsDefault reads an environment variable and returns its value or a default string.
func getEnvAsDefault(key string, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func normalizeS3Prefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "Sensio/"
	}

	prefix = strings.TrimLeft(prefix, "/")
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return prefix
}

func validateS3RequiredEnv(cfg *Config) {
	if cfg == nil || !cfg.S3Enabled || isDevEnvironment(cfg.ApplicationEnvironment) {
		return
	}

	missing := make([]string, 0, 4)
	if strings.TrimSpace(cfg.S3Bucket) == "" {
		missing = append(missing, "S3_BUCKET")
	}
	if strings.TrimSpace(cfg.S3Region) == "" {
		missing = append(missing, "S3_REGION")
	}
	if strings.TrimSpace(cfg.AWSAccessKeyID) == "" {
		missing = append(missing, "AWS_ACCESS_KEY_ID")
	}
	if strings.TrimSpace(cfg.AWSSecretAccessKey) == "" {
		missing = append(missing, "AWS_SECRET_ACCESS_KEY")
	}

	if len(missing) > 0 {
		log.Fatalf("Config: S3 is enabled in non-dev environment but required env vars are missing: %s", strings.Join(missing, ", "))
	}

	if cfg.S3SignedURLTTLSeconds <= 0 {
		log.Fatalf("Config: S3_SIGNED_URL_TTL_SECONDS must be greater than 0")
	}
}

func isDevEnvironment(env string) bool {
	normalized := strings.ToLower(strings.TrimSpace(env))
	if normalized == "" {
		return true
	}

	return normalized == "dev" || normalized == "development" || normalized == "local" || normalized == "test"
}

// getEnvDuration reads an environment variable and ensures it's a valid Go duration and NOT EMPTY.
// It fails fast (log.Fatalf) if the variable is missing or invalid.
func getEnvDuration(key string) string {
	val, err := validateEnvDuration(key)
	if err != nil {
		log.Fatalf("Config: %v", err)
	}
	return val
}

// validateEnvDuration checks if an environment variable is a valid Go duration and NOT EMPTY.
// It returns the value or an error. This is separated for unit testing.
func validateEnvDuration(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("missing required environment variable %s (expected Go duration format like '8h', '30m', '24h')", key)
	}
	if _, err := time.ParseDuration(val); err != nil {
		return "", fmt.Errorf("invalid duration format for %s: '%s' - %v (expected e.g. '8h', '30m', '24h')", key, val, err)
	}
	return val, nil
}
