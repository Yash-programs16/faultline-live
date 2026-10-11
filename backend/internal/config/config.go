package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Ollama   OllamaConfig
	Faults   FaultsConfig
	Telemetry TelemetryConfig
}

type ServerConfig struct {
	Host string
	Port string
}

func (s ServerConfig) Addr() string {
	return s.Host + ":" + s.Port
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

type AuthConfig struct {
	JWTSecret     string
	TokenExpiry   time.Duration
	BCryptCost    int
}

type OllamaConfig struct {
	BaseURL          string
	Model            string
	Timeout          time.Duration
	MaxRetries       int
	Enabled          bool
}

type FaultsConfig struct {
	Enabled       bool
	MaxDuration   time.Duration
	MaxLatencyMs  int
	MaxErrorRate  float64
}

type TelemetryConfig struct {
	MetricsWindow    time.Duration
	IncidentCooldown time.Duration
	SampleRetention  time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL:             getRequiredEnv("DATABASE_URL"),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 20),
			MinConns:        getEnvInt("DB_MIN_CONNS", 2),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:  getEnvDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:   getRequiredEnv("JWT_SECRET"),
			TokenExpiry: getEnvDuration("JWT_TOKEN_EXPIRY", 24*time.Hour),
			BCryptCost:  getEnvInt("BCRYPT_COST", 12),
		},
		Ollama: OllamaConfig{
			BaseURL:    getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			Model:      getEnv("OLLAMA_MODEL", "gemma2:2b"),
			Timeout:    getEnvDuration("OLLAMA_TIMEOUT", 60*time.Second),
			MaxRetries: getEnvInt("OLLAMA_MAX_RETRIES", 2),
			Enabled:    getEnvBool("OLLAMA_ENABLED", true),
		},
		Faults: FaultsConfig{
			Enabled:      getEnvBool("FAULTS_ENABLED", false),
			MaxDuration:  getEnvDuration("FAULTS_MAX_DURATION", 10*time.Minute),
			MaxLatencyMs: getEnvInt("FAULTS_MAX_LATENCY_MS", 10000),
			MaxErrorRate: getEnvFloat("FAULTS_MAX_ERROR_RATE", 0.5),
		},
		Telemetry: TelemetryConfig{
			MetricsWindow:    getEnvDuration("TELEMETRY_WINDOW", 5*time.Minute),
			IncidentCooldown: getEnvDuration("INCIDENT_COOLDOWN", 2*time.Minute),
			SampleRetention:  getEnvDuration("SAMPLE_RETENTION", 24*time.Hour),
		},
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getRequiredEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable " + key + " is not set")
	}
	return v
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}