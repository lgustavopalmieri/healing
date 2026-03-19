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
			MaxOpenConns:    getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("POSTGRES_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		Kafka: KafkaConfig{
			BootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", ""),
			AutoOffsetReset:  getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses:    getEnvAsSlice("ELASTICSEARCH_ADDRESSES", nil),
			MaxRetries:   getEnvAsInt("ELASTICSEARCH_MAX_RETRIES", 3),
			RetryBackoff: getEnvAsDuration("ELASTICSEARCH_RETRY_BACKOFF", 1*time.Second),
		},
		External: ExternalConfig{
			LicenseBaseURL: getEnv("LICENSE_VALIDATION_BASE_URL", ""),
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
