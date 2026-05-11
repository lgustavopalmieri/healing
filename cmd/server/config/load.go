package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	env := getEnv("APP_ENV", "production")

	_ = loadEnvFile(env)

	cfg := &Config{
		Server: ServerConfig{
			GRPCPort:          getEnvAsInt("SERVER_GRPC_PORT", 50051),
			HTTPPort:          getEnvAsInt("SERVER_HTTP_PORT", 8080),
			ShutdownTimeout:   getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			MaxConnections:    getEnvAsInt("SERVER_MAX_CONNECTIONS", 1000),
			ConnectionTimeout: getEnvAsDuration("SERVER_CONNECTION_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("POSTGRES_HOST", ""),
			Port:            getEnvAsInt("POSTGRES_PORT", 5432),
			User:            getEnv("POSTGRES_USER", ""),
			Password:        getEnv("POSTGRES_PASSWORD", ""),
			Database:        getEnv("POSTGRES_DB", ""),
			SSLMode:         getEnv("POSTGRES_SSLMODE", "require"),
			MaxOpenConns:    getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("POSTGRES_CONN_MAX_IDLE_TIME", 2*time.Minute),
		},
		AuthDB: DatabaseConfig{
			Host:            getEnv("AUTH_POSTGRES_HOST", ""),
			Port:            getEnvAsInt("AUTH_POSTGRES_PORT", 5432),
			User:            getEnv("AUTH_POSTGRES_USER", ""),
			Password:        getEnv("AUTH_POSTGRES_PASSWORD", ""),
			Database:        getEnv("AUTH_POSTGRES_DB", ""),
			SSLMode:         getEnv("AUTH_POSTGRES_SSLMODE", "require"),
			MaxOpenConns:    getEnvAsInt("AUTH_POSTGRES_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvAsInt("AUTH_POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("AUTH_POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("AUTH_POSTGRES_CONN_MAX_IDLE_TIME", 1*time.Minute),
		},
		SQS: SQSConfig{
			Region:      getEnv("SQS_REGION", ""),
			QueuePrefix: getEnv("SQS_QUEUE_PREFIX", "specialist"),
			Endpoint:    getEnv("SQS_ENDPOINT", ""),
		},
		SNS: SNSConfig{
			Region:      getEnv("SNS_REGION", getEnv("SQS_REGION", "")),
			Endpoint:    getEnv("SNS_ENDPOINT", ""),
			TopicPrefix: getEnv("SNS_TOPIC_PREFIX", getEnv("SQS_QUEUE_PREFIX", "specialist")),
		},
		OpenSearch: OpenSearchConfig{
			Addresses:   getEnvAsSlice("OPENSEARCH_ADDRESSES", nil),
			Region:      getEnv("OPENSEARCH_REGION", ""),
			IndexPrefix: getEnv("OPENSEARCH_INDEX_PREFIX", ""),
		},
		External: ExternalConfig{
			LicenseBaseURL: getEnv("LICENSE_VALIDATION_BASE_URL", ""),
		},
		Email: EmailConfig{
			SMTPHost:    getEnv("EMAIL_SMTP_HOST", "mailhog"),
			SMTPPort:    getEnvAsInt("EMAIL_SMTP_PORT", 1025),
			FromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@healing.local"),
			FromName:    getEnv("EMAIL_FROM_NAME", "Healing Platform"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", ""),
			Port:         getEnvAsInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
		},
		Auth: AuthConfig{
			PrivateKeyPath:    getEnv("AUTH_PRIVATE_KEY_PATH", ""),
			PublicKeyPath:     getEnv("AUTH_PUBLIC_KEY_PATH", ""),
			CurrentKeyID:      getEnv("AUTH_CURRENT_KEY_ID", ""),
			AccessTokenTTL:    getEnvAsDuration("AUTH_ACCESS_TOKEN_TTL", 1*time.Hour),
			RefreshTokenTTL:   getEnvAsDuration("AUTH_REFRESH_TOKEN_TTL", 168*time.Hour),
			SetPasswordTTL:    getEnvAsDuration("AUTH_SET_PASSWORD_TTL", 24*time.Hour),
			ResetPasswordTTL:  getEnvAsDuration("AUTH_RESET_PASSWORD_TTL", 1*time.Hour),
			Issuer:            getEnv("AUTH_ISSUER", "healing-specialist"),
			Audience:          getEnv("AUTH_AUDIENCE", "healing-platform"),
			BcryptCost:        getEnvAsInt("AUTH_BCRYPT_COST", 12),
			PasswordMinLength: getEnvAsInt("AUTH_PASSWORD_MIN_LENGTH", 8),
		},
		Otel: OtelConfig{
			ExporterEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
			ExporterProtocol:   getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf"),
			ServiceName:        getEnv("OTEL_SERVICE_NAME", "healing-specialist"),
			ResourceAttributes: getEnv("OTEL_RESOURCE_ATTRIBUTES", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	log.Printf("Configuration loaded (env=%s)", env)

	return cfg, nil
}

func loadEnvFile(env string) error {
	envDir := getEnv("ENV_DIR", "")

	if envDir == "" {
		if _, err := os.Stat(".env"); err == nil {
			envDir = "."
		}
	}

	if envDir == "" {
		return nil
	}

	var envFile string
	switch env {
	case "test":
		envFile = filepath.Join(envDir, ".env.test")
	case "production":
		envFile = filepath.Join(envDir, ".env.production")
	default:
		envFile = filepath.Join(envDir, ".env")
	}

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return nil
	}

	viper.SetConfigFile(envFile)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file %s: %w", envFile, err)
	}

	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("error setting env var %s: %w", key, err)
			}
		}
	}

	log.Printf("Loaded env file: %s", envFile)
	return nil
}
