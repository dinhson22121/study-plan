// Package config loads application configuration from a YAML file overlaid with
// environment variables (env wins). All runtime wiring reads from Config.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the fully-resolved application configuration.
type Config struct {
	Env      string         `mapstructure:"env"`
	Port     string         `mapstructure:"port"`
	Timezone string         `mapstructure:"timezone"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	FCM      FCMConfig      `mapstructure:"fcm"`
	S3       S3Config       `mapstructure:"s3"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

// S3Config configures the S3-compatible object store (MinIO in local dev).
type S3Config struct {
	Endpoint     string `mapstructure:"endpoint"`
	Region       string `mapstructure:"region"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	Bucket       string `mapstructure:"bucket"`
	UsePathStyle bool   `mapstructure:"use_path_style"`
}

// UploadConfig configures admin asset uploads.
type UploadConfig struct {
	MaxFileSizeBytes int64         `mapstructure:"max_file_size_bytes"`
	PresignTTL       time.Duration `mapstructure:"presign_ttl"`
}

type PostgresConfig struct {
	URL          string `mapstructure:"url"`
	MaxConns     int32  `mapstructure:"max_conns"`
	MinConns     int32  `mapstructure:"min_conns"`
	MigrationDir string `mapstructure:"migration_dir"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type KafkaConfig struct {
	Brokers    []string `mapstructure:"brokers"`
	GroupID    string   `mapstructure:"group_id"`
	Partitions int      `mapstructure:"partitions"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
	Issuer     string        `mapstructure:"issuer"`
}

type FCMConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	ProjectID       string `mapstructure:"project_id"`
}

// Load reads configuration from config.yaml (searched in path, the current dir,
// and ./config) plus EDU_* environment variables, then validates required
// fields. Env vars use underscores for nesting, e.g. EDU_POSTGRES_URL.
func Load(paths ...string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	for _, p := range paths {
		v.AddConfigPath(p)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	setDefaults(v)

	v.SetEnvPrefix("EDU")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindEnvs(v)

	if err := v.ReadInConfig(); err != nil {
		// A missing file is fine when env vars supply everything; other read
		// errors (malformed YAML) are fatal.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// bindEnvs explicitly binds every key to its EDU_* env var. AutomaticEnv alone
// only resolves keys viper already knows (from defaults or a loaded file), so
// nested keys supplied solely via env (postgres.url, jwt.secret, ...) need an
// explicit binding to be picked up by Unmarshal.
func bindEnvs(v *viper.Viper) {
	keys := []string{
		"env", "port", "timezone",
		"postgres.url", "postgres.max_conns", "postgres.min_conns", "postgres.migration_dir",
		"redis.url",
		"kafka.brokers", "kafka.group_id", "kafka.partitions",
		"jwt.secret", "jwt.access_ttl", "jwt.refresh_ttl", "jwt.issuer",
		"fcm.credentials_file", "fcm.project_id",
		"s3.endpoint", "s3.region", "s3.access_key", "s3.secret_key", "s3.bucket",
		"upload.presign_ttl",
	}
	for _, k := range keys {
		_ = v.BindEnv(k)
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "development")
	v.SetDefault("port", ":8080")
	v.SetDefault("timezone", "Asia/Ho_Chi_Minh")
	v.SetDefault("postgres.max_conns", 10)
	v.SetDefault("postgres.min_conns", 2)
	v.SetDefault("postgres.migration_dir", "migrations")
	v.SetDefault("kafka.group_id", "edu-app")
	v.SetDefault("kafka.partitions", 1)
	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 720*time.Hour) // 30 days
	v.SetDefault("jwt.issuer", "edu-app")
	v.SetDefault("s3.region", "us-east-1")
	v.SetDefault("s3.use_path_style", true)                  // MinIO requires path-style addressing
	v.SetDefault("upload.max_file_size_bytes", 20*1024*1024) // 20MB
	v.SetDefault("upload.presign_ttl", 15*time.Minute)
}

// validate enforces that secrets and connection strings required to boot are
// present, failing fast with a clear message (no silent zero-value startup).
func (c *Config) validate() error {
	var missing []string
	if c.Postgres.URL == "" {
		missing = append(missing, "postgres.url (EDU_POSTGRES_URL)")
	}
	if c.Redis.URL == "" {
		missing = append(missing, "redis.url (EDU_REDIS_URL)")
	}
	if len(c.Kafka.Brokers) == 0 {
		missing = append(missing, "kafka.brokers (EDU_KAFKA_BROKERS)")
	}
	if c.JWT.Secret == "" {
		missing = append(missing, "jwt.secret (EDU_JWT_SECRET)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}
	return nil
}
