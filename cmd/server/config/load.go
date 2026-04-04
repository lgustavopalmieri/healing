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
		SQS: SQSConfig{
			Region:      getEnv("SQS_REGION", ""),
			QueuePrefix: getEnv("SQS_QUEUE_PREFIX", "specialist"),
			Endpoint:    getEnv("SQS_ENDPOINT", ""),
		},
		OpenSearch: OpenSearchConfig{
			Addresses:   getEnvAsSlice("OPENSEARCH_ADDRESSES", nil),
			Region:      getEnv("OPENSEARCH_REGION", ""),
			IndexPrefix: getEnv("OPENSEARCH_INDEX_PREFIX", ""),
		},
		External: ExternalConfig{
			LicenseBaseURL: getEnv("LICENSE_VALIDATION_BASE_URL", ""),
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
