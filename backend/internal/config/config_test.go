package config

import (
	"os"
	"testing"
)

func setupEnv(t *testing.T) {
	t.Helper()
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/test")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("FAULT_MAX_LATENCY_MS", "10000")
	os.Setenv("FAULT_MAX_FAILURE_RATE", "1.0")
	os.Setenv("FAULT_MAX_DURATION_SECONDS", "300")
	os.Setenv("FAULT_LOCAL_ONLY", "true")
	os.Setenv("TELEMETRY_WINDOW_SIZE", "5s")
	os.Setenv("TELEMETRY_BUFFER_SIZE", "120")
	os.Setenv("TELEMETRY_PERSIST_ENABLED", "true")
}

func TestLoad_ValidConfig(t *testing.T) {
	setupEnv(t)
	os.Setenv("FAULT_MAX_LATENCY_MS", "5000")
	os.Setenv("FAULT_MAX_FAILURE_RATE", "0.5")
	os.Setenv("FAULT_MAX_DURATION_SECONDS", "120")
	os.Setenv("FAULT_LOCAL_ONLY", "true")
	os.Setenv("TELEMETRY_WINDOW_SIZE", "5s")
	os.Setenv("TELEMETRY_BUFFER_SIZE", "60")
	os.Setenv("TELEMETRY_PERSIST_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Faults.MaxLatencyMs != 5000 {
		t.Errorf("expected max latency 5000, got %d", cfg.Faults.MaxLatencyMs)
	}
	if cfg.Faults.MaxFailureRate != 0.5 {
		t.Errorf("expected max failure rate 0.5, got %f", cfg.Faults.MaxFailureRate)
	}
	if cfg.Faults.MaxDurationSeconds != 120 {
		t.Errorf("expected max duration 120, got %d", cfg.Faults.MaxDurationSeconds)
	}
	if !cfg.Faults.LocalOnly {
		t.Error("expected LocalOnly true")
	}
	if cfg.Telemetry.WindowSize.String() != "5s" {
		t.Errorf("expected window size 5s, got %s", cfg.Telemetry.WindowSize)
	}
	if cfg.Telemetry.BufferSize != 60 {
		t.Errorf("expected buffer size 60, got %d", cfg.Telemetry.BufferSize)
	}
	if !cfg.Telemetry.PersistEnabled {
		t.Error("expected persist enabled true")
	}
}

func TestLoad_Defaults(t *testing.T) {
	setupEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Faults.MaxLatencyMs != 10000 {
		t.Errorf("expected default max latency 10000, got %d", cfg.Faults.MaxLatencyMs)
	}
	if cfg.Faults.MaxFailureRate != 1.0 {
		t.Errorf("expected default max failure rate 1.0, got %f", cfg.Faults.MaxFailureRate)
	}
	if cfg.Faults.MaxDurationSeconds != 300 {
		t.Errorf("expected default max duration 300, got %d", cfg.Faults.MaxDurationSeconds)
	}
	if !cfg.Faults.LocalOnly {
		t.Error("expected default LocalOnly true")
	}
	if cfg.Telemetry.WindowSize.String() != "5s" {
		t.Errorf("expected default window size 5s, got %s", cfg.Telemetry.WindowSize)
	}
	if cfg.Telemetry.BufferSize != 120 {
		t.Errorf("expected default buffer size 120, got %d", cfg.Telemetry.BufferSize)
	}
	if !cfg.Telemetry.PersistEnabled {
		t.Error("expected default persist enabled true")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	setupEnv(t)
	os.Setenv("SERVER_PORT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoad_InvalidWindowSize(t *testing.T) {
	setupEnv(t)
	os.Setenv("TELEMETRY_WINDOW_SIZE", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid window size")
	}
}

func TestLoad_InvalidFailureRate(t *testing.T) {
	setupEnv(t)
	os.Setenv("FAULT_MAX_FAILURE_RATE", "1.5")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for failure rate > 1.0")
	}
}

func TestLoad_InvalidMaxLatency(t *testing.T) {
	setupEnv(t)
	os.Setenv("FAULT_MAX_LATENCY_MS", "50")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for max latency < 100")
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	setupEnv(t)
	os.Setenv("FAULT_MAX_DURATION_SECONDS", "5")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for duration < 10")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("FAULT_MAX_LATENCY_MS", "10000")
	os.Setenv("FAULT_MAX_FAILURE_RATE", "1.0")
	os.Setenv("FAULT_MAX_DURATION_SECONDS", "300")
	os.Setenv("FAULT_LOCAL_ONLY", "true")
	os.Setenv("TELEMETRY_WINDOW_SIZE", "5s")
	os.Setenv("TELEMETRY_BUFFER_SIZE", "120")
	os.Setenv("TELEMETRY_PERSIST_ENABLED", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
	if err != ErrMissingDatabaseURL {
		t.Errorf("expected ErrMissingDatabaseURL, got %v", err)
	}
}