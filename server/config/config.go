package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env       string          `mapstructure:"env"`
	Port      string          `mapstructure:"port"`
	Timezone  string          `mapstructure:"timezone"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Kafka     KafkaConfig     `mapstructure:"kafka"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	FCM       FCMConfig       `mapstructure:"fcm"`
	S3        S3Config        `mapstructure:"s3"`
	Upload    UploadConfig    `mapstructure:"upload"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
}

type RateLimitConfig struct {
	AuthRequests int           `mapstructure:"auth_requests"`
	AuthWindow   time.Duration `mapstructure:"auth_window"`
}

type S3Config struct {
	Endpoint     string `mapstructure:"endpoint"`
	Region       string `mapstructure:"region"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	Bucket       string `mapstructure:"bucket"`
	UsePathStyle bool   `mapstructure:"use_path_style"`
}

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

func bindEnvs(v *viper.Viper) {
	keys := []string{
		"env", "port", "timezone",
		"postgres.url", "postgres.max_conns", "postgres.min_conns", "postgres.migration_dir",
		"redis.url",
		"kafka.brokers", "kafka.group_id", "kafka.partitions",
		"jwt.secret", "jwt.access_ttl", "jwt.refresh_ttl", "jwt.issuer",
		"fcm.credentials_file", "fcm.project_id",
		"s3.endpoint", "s3.region", "s3.access_key", "s3.secret_key", "s3.bucket", "s3.use_path_style",
		"upload.max_file_size_bytes", "upload.presign_ttl",
		"ratelimit.auth_requests", "ratelimit.auth_window",
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
	v.SetDefault("jwt.refresh_ttl", 720*time.Hour)
	v.SetDefault("jwt.issuer", "edu-app")
	v.SetDefault("s3.region", "us-east-1")
	v.SetDefault("s3.use_path_style", true)
	v.SetDefault("upload.max_file_size_bytes", 20*1024*1024)
	v.SetDefault("upload.presign_ttl", 15*time.Minute)
	v.SetDefault("ratelimit.auth_requests", 10)
	v.SetDefault("ratelimit.auth_window", time.Minute)
}

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
	if strings.EqualFold(c.Env, "production") {
		return c.validateProduction()
	}
	return nil
}

const minProdSecretLen = 32

var weakSecrets = map[string]bool{
	"dev-only-change-me-in-production": true,
	"change-me-in-production":          true,
	"dev-only-change-me":               true,
}

func (c *Config) validateProduction() error {
	var problems []string
	if weakSecrets[c.JWT.Secret] || len(c.JWT.Secret) < minProdSecretLen {
		problems = append(problems, fmt.Sprintf("jwt.secret must be a strong non-default value of at least %d characters", minProdSecretLen))
	}
	if !postgresUsesTLS(c.Postgres.URL) {
		problems = append(problems, "postgres.url must use sslmode=require (or stricter) in production")
	}
	if len(problems) > 0 {
		return fmt.Errorf("insecure production config: %s", strings.Join(problems, "; "))
	}
	return nil
}

func postgresUsesTLS(url string) bool {
	lower := strings.ToLower(url)
	for _, mode := range []string{"sslmode=require", "sslmode=verify-ca", "sslmode=verify-full"} {
		if strings.Contains(lower, mode) {
			return true
		}
	}
	return false
}
