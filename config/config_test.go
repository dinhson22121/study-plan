package config

import (
	"os"
	"testing"
	"time"
)

// chdirTemp moves into an empty working directory so Load's implicit "." and
// "./config" search paths don't pick up the repo's dev config.yaml, keeping
// these tests hermetic. Restores the original cwd on cleanup.
func chdirTemp(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestLoad_EnvOverridesAndDefaults(t *testing.T) {
	// Arrange: supply required fields via env; rely on defaults for the rest.
	chdirTemp(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x")
	t.Setenv("EDU_REDIS_URL", "redis://cache:6379/0")
	t.Setenv("EDU_KAFKA_BROKERS", "b1:9092,b2:9092")
	t.Setenv("EDU_JWT_SECRET", "supersecret")

	// Act
	cfg, err := Load()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Postgres.URL != "postgres://u:p@db:5432/x" {
		t.Fatalf("env did not override postgres url: %q", cfg.Postgres.URL)
	}
	if len(cfg.Kafka.Brokers) != 2 || cfg.Kafka.Brokers[0] != "b1:9092" {
		t.Fatalf("env brokers not parsed as slice: %v", cfg.Kafka.Brokers)
	}
	if cfg.JWT.AccessTTL != 15*time.Minute {
		t.Fatalf("expected default access ttl 15m, got %v", cfg.JWT.AccessTTL)
	}
	if cfg.Port != ":8080" {
		t.Fatalf("expected default port :8080, got %q", cfg.Port)
	}
}

func TestLoad_FailsWhenRequiredMissing(t *testing.T) {
	// Only set some required fields; jwt.secret missing should fail validation.
	chdirTemp(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x")
	t.Setenv("EDU_REDIS_URL", "redis://cache:6379/0")
	t.Setenv("EDU_KAFKA_BROKERS", "b1:9092")

	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing jwt.secret")
	}
}
