package config

import (
	"testing"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	for k, v := range map[string]string{
		"DATABASE_URL":     "postgres://user:pass@localhost/test",
		"JWT_SECRET":       "0123456789abcdef0123456789abcdef",
		"SERVER_PORT":      "8080",
		"FAULTS_ENABLED":   "false",
		"TELEMETRY_WINDOW": "5s",
	} {
		t.Setenv(k, v)
	}
}

func TestLoadValidConfig(t *testing.T) {
	setValidEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "8080" {
		t.Fatalf("port = %q", cfg.Server.Port)
	}
	if cfg.Telemetry.MetricsWindow.String() != "5s" {
		t.Fatalf("window = %s", cfg.Telemetry.MetricsWindow)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
}

func TestLoadRequiresStrongJWTSecret(t *testing.T) {
	setValidEnv(t)
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("expected weak JWT_SECRET error")
	}
}

func TestMalformedDurationUsesDefault(t *testing.T) {
	setValidEnv(t)
	t.Setenv("TELEMETRY_WINDOW", "invalid")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Telemetry.MetricsWindow <= 0 {
		t.Fatal("expected positive default metrics window")
	}
}
