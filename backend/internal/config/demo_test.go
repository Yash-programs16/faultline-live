package config

import (
	"os"
	"testing"
)

func TestDemoModeConfig(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://neondb_owner:npg_9LvgVyZuYz1b@ep-broad-fire-b5sov2sg-pooler.c-7.us-east-2.aws.neon.tech/neondb?sslmode=require&channel_binding=require")
	os.Setenv("JWT_SECRET", "nUe2tz3GMapU8EQ73SMM_3zlks7VWeRi2nvhMjSB6VDH4m0-YZrum1mc0xNKyZS9")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("FAULTLINE_DEMO_MODE", "true")
	os.Setenv("FAULTS_ENABLED", "true")
	os.Setenv("TELEMETRY_WINDOW", "1m")
	os.Setenv("INCIDENT_COOLDOWN", "30s")
	
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("DemoMode: %v", cfg.DemoMode)
	t.Logf("OpsAuthRequired: %v", cfg.OpsAuthRequired)
	t.Logf("Faults.Enabled: %v", cfg.Faults.Enabled)
	t.Logf("Server.Host: %v", cfg.Server.Host)
	t.Logf("Server.Port: %v", cfg.Server.Port)
	
	if !cfg.DemoMode {
		t.Error("Expected DemoMode=true")
	}
	if cfg.OpsAuthRequired {
		t.Error("Expected OpsAuthRequired=false in demo mode")
	}
	if !cfg.Faults.Enabled {
		t.Error("Expected Faults.Enabled=true")
	}
}
