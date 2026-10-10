package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Telemetry TelemetryConfig
	Faults    FaultsConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	URL string
}

type TelemetryConfig struct {
	WindowSize     time.Duration
	BufferSize     int
	PersistEnabled bool
}

type FaultsConfig struct {
	MaxLatencyMs      int
	MaxFailureRate    float64
	MaxDurationSeconds int
	LocalOnly         bool
}

func Load() (*Config, error) {
	port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	maxLatency, _ := strconv.Atoi(getEnv("FAULT_MAX_LATENCY_MS", "10000"))
	maxFailureRate, _ := strconv.ParseFloat(getEnv("FAULT_MAX_FAILURE_RATE", "1.0"), 64)
	maxDuration, _ := strconv.Atoi(getEnv("FAULT_MAX_DURATION_SECONDS", "300"))
	localOnly := getEnv("FAULT_LOCAL_ONLY", "true") == "true"

	windowSize, _ := time.ParseDuration(getEnv("TELEMETRY_WINDOW_SIZE", "5s"))
	bufferSize, _ := strconv.Atoi(getEnv("TELEMETRY_BUFFER_SIZE", "120"))
	persistEnabled := getEnv("TELEMETRY_PERSIST_ENABLED", "true") == "true"

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, ErrMissingDatabaseURL
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: port,
		},
		Database: DatabaseConfig{
			URL: dbURL,
		},
		Telemetry: TelemetryConfig{
			WindowSize:     windowSize,
			BufferSize:     bufferSize,
			PersistEnabled: persistEnabled,
		},
		Faults: FaultsConfig{
			MaxLatencyMs:       maxLatency,
			MaxFailureRate:     maxFailureRate,
			MaxDurationSeconds: maxDuration,
			LocalOnly:          localOnly,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var ErrMissingDatabaseURL = &ConfigError{"DATABASE_URL environment variable is required"}

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}