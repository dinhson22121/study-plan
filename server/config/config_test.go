package config

import (
	"os"
	"testing"
	"time"
)

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

	chdirTemp(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x")
	t.Setenv("EDU_REDIS_URL", "redis://cache:6379/0")
	t.Setenv("EDU_KAFKA_BROKERS", "b1:9092,b2:9092")
	t.Setenv("EDU_JWT_SECRET", "supersecret")
	t.Setenv("EDU_S3_USE_PATH_STYLE", "false")
	t.Setenv("EDU_UPLOAD_MAX_FILE_SIZE_BYTES", "4096")

	cfg, err := Load()

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
	if cfg.S3.UsePathStyle {
		t.Fatalf("expected s3.use_path_style override to be false")
	}
	if cfg.Upload.MaxFileSizeBytes != 4096 {
		t.Fatalf("expected upload.max_file_size_bytes override to be 4096, got %d", cfg.Upload.MaxFileSizeBytes)
	}
}

func TestLoad_FailsWhenRequiredMissing(t *testing.T) {

	chdirTemp(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x")
	t.Setenv("EDU_REDIS_URL", "redis://cache:6379/0")
	t.Setenv("EDU_KAFKA_BROKERS", "b1:9092")

	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing jwt.secret")
	}
}

func setBaseProdEnv(t *testing.T) {
	t.Helper()
	chdirTemp(t)
	t.Setenv("EDU_ENV", "production")
	t.Setenv("EDU_REDIS_URL", "redis://cache:6379/0")
	t.Setenv("EDU_KAFKA_BROKERS", "b1:9092")
}

func TestLoad_ProductionRejectsWeakSecret(t *testing.T) {
	setBaseProdEnv(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x?sslmode=require")
	t.Setenv("EDU_JWT_SECRET", "dev-only-change-me-in-production")

	if _, err := Load(); err == nil {
		t.Fatalf("expected error for weak production jwt secret")
	}
}

func TestLoad_ProductionRejectsShortSecret(t *testing.T) {
	setBaseProdEnv(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x?sslmode=require")
	t.Setenv("EDU_JWT_SECRET", "tooshort")

	if _, err := Load(); err == nil {
		t.Fatalf("expected error for short production jwt secret")
	}
}

func TestLoad_ProductionRejectsInsecurePostgres(t *testing.T) {
	setBaseProdEnv(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x?sslmode=disable")
	t.Setenv("EDU_JWT_SECRET", "a-sufficiently-long-production-secret-value")

	if _, err := Load(); err == nil {
		t.Fatalf("expected error for insecure postgres sslmode in production")
	}
}

func TestLoad_ProductionAcceptsStrongConfig(t *testing.T) {
	setBaseProdEnv(t)
	t.Setenv("EDU_POSTGRES_URL", "postgres://u:p@db:5432/x?sslmode=require")
	t.Setenv("EDU_JWT_SECRET", "a-sufficiently-long-production-secret-value")

	if _, err := Load(); err != nil {
		t.Fatalf("unexpected error for valid production config: %v", err)
	}
}
