package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server          ServerConfig
	Database        DatabaseConfig
	Auth            AuthConfig
	OpsAuthRequired bool
	DemoMode        bool
	AllowPublicDemo bool
	OperatorIDs     string
	CORSOrigins     []string
	AutoMigrate     bool
	Ollama          OllamaConfig
	Gemini          GeminiConfig
	Faults          FaultsConfig
	Telemetry       TelemetryConfig
}

type ServerConfig struct {
	Host string
	Port string
}

func (s ServerConfig) Addr() string {
	return net.JoinHostPort(s.Host, s.Port)
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
	JWTSecret   string
	TokenExpiry time.Duration
	BCryptCost  int
}

type OllamaConfig struct {
	BaseURL    string
	Model      string
	Timeout    time.Duration
	MaxRetries int
	Enabled    bool
}

// GeminiConfig drives cloud investigations via the Generative Language API.
// APIKey must be provided via the GEMINI_API_KEY environment variable and
// must never be committed to the repository.
type GeminiConfig struct {
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
	Enabled    bool
}

type FaultsConfig struct {
	Enabled      bool
	MaxDuration  time.Duration
	MaxLatencyMs int
	MaxErrorRate float64
}

type TelemetryConfig struct {
	MetricsWindow    time.Duration
	IncidentCooldown time.Duration
	SampleRetention  time.Duration
}

var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	if databaseURL == "" {
		return nil, ErrMissingDatabaseURL
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be set and contain at least 32 characters")
	}
	// Render (and most PaaS hosts) inject PORT. It takes precedence over
	// SERVER_PORT so the same binary runs locally and in the cloud.
	port := getEnv("SERVER_PORT", "8080")
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "127.0.0.1"),
			Port: port,
		},
		Database: DatabaseConfig{
			URL:             databaseURL,
			MaxConns:        getEnvInt("DB_MAX_CONNS", 20),
			MinConns:        getEnvInt("DB_MIN_CONNS", 2),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:  getEnvDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
		},
		OpsAuthRequired: getEnvBool("OPS_AUTH_REQUIRED", true),
		DemoMode:        getEnvBool("FAULTLINE_DEMO_MODE", false),
		AllowPublicDemo: getEnvBool("ALLOW_PUBLIC_DEMO", false),
		OperatorIDs:     os.Getenv("OPS_OPERATOR_IDS"),
		CORSOrigins:     getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		AutoMigrate:     getEnvBool("AUTO_MIGRATE", false),
		Auth: AuthConfig{
			JWTSecret:   jwtSecret,
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
		Gemini: GeminiConfig{
			APIKey:     os.Getenv("GEMINI_API_KEY"),
			Model:      getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
			Timeout:    getEnvDuration("GEMINI_TIMEOUT", 45*time.Second),
			MaxRetries: getEnvInt("GEMINI_MAX_RETRIES", 1),
			Enabled:    getEnvBool("GEMINI_ENABLED", true),
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
	if cfg.Server.Port == "" {
		return nil, fmt.Errorf("SERVER_PORT must not be empty")
	}
	loopback := net.ParseIP(cfg.Server.Host)
	if cfg.DemoMode {
		// Local demo stays loopback-only unless the operator explicitly opts
		// into a public demo (hosted environments like Render must bind
		// 0.0.0.0). Public demo mode is insecure by design: anyone with the
		// URL can drive simulations, so only use it for throwaway demos.
		if loopback == nil || !loopback.IsLoopback() {
			if !cfg.AllowPublicDemo || strings.EqualFold(os.Getenv("APP_ENV"), "production") || strings.EqualFold(os.Getenv("GO_ENV"), "production") {
				return nil, fmt.Errorf("FAULTLINE_DEMO_MODE with a non-loopback listener requires ALLOW_PUBLIC_DEMO=true and a non-production environment")
			}
		} else if strings.EqualFold(os.Getenv("APP_ENV"), "production") || strings.EqualFold(os.Getenv("GO_ENV"), "production") {
			return nil, fmt.Errorf("FAULTLINE_DEMO_MODE requires a non-production environment")
		}
		cfg.OpsAuthRequired = false
	} else if !cfg.OpsAuthRequired {
		return nil, fmt.Errorf("unauthenticated operations require explicit FAULTLINE_DEMO_MODE=true")
	}
	// Merge FRONTEND_URL into CORS origins so a single env var is enough on
	// hosted platforms (Render sets it, Vercel consumes it and vice versa).
	if fu := strings.TrimSpace(os.Getenv("FRONTEND_URL")); fu != "" {
		found := false
		for _, o := range cfg.CORSOrigins {
			if strings.TrimRight(o, "/") == strings.TrimRight(fu, "/") {
				found = true
				break
			}
		}
		if !found {
			cfg.CORSOrigins = append(cfg.CORSOrigins, fu)
		}
	}
	if cfg.Faults.Enabled && !cfg.DemoMode {
		return nil, fmt.Errorf("FAULTS_ENABLED requires FAULTLINE_DEMO_MODE=true")
	}

	if cfg.Telemetry.MetricsWindow <= 0 {
		return nil, fmt.Errorf("TELEMETRY_WINDOW must be positive")
	}
	if cfg.Faults.MaxDuration <= 0 || cfg.Faults.MaxLatencyMs < 0 || cfg.Faults.MaxErrorRate < 0 || cfg.Faults.MaxErrorRate > 1 {
		return nil, fmt.Errorf("invalid fault configuration")
	}
	if cfg.Database.MaxConns < 1 || cfg.Database.MinConns < 0 || cfg.Database.MinConns > cfg.Database.MaxConns {
		return nil, fmt.Errorf("invalid database pool configuration")
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvList(key string, defaultValue []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return defaultValue
	}
	return out
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
