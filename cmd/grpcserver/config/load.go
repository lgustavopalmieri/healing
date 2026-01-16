package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	// Determine environment (default: development)
	env := getEnv("APP_ENV", "development")

	// Load environment-specific .env file
	if err := loadEnvFile(env); err != nil {
		return nil, fmt.Errorf("failed to load environment file: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			GRPCPort:          getEnvAsInt("SERVER_GRPC_PORT", 50051),
			ShutdownTimeout:   getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			MaxConnections:    getEnvAsInt("SERVER_MAX_CONNECTIONS", 1000),
			ConnectionTimeout: getEnvAsDuration("SERVER_CONNECTION_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Database: getEnv("POSTGRES_DB", ""),
		},
		Kafka: KafkaConfig{
			BootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", ""),
			AutoOffsetReset:  getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
		},
		Observability: ObservabilityConfig{
			ServiceName:    getEnv("OTEL_SERVICE_NAME", ""),
			ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
			Environment:    getEnv("OTEL_ENVIRONMENT", env),
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_GRPC_ENDPOINT", ""),
			OTLPProtocol:   getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadEnvFile loads the appropriate .env file based on the environment
func loadEnvFile(env string) error {
	// Determine the config directory (cmd/grpcserver)
	configDir := "cmd/grpcserver"

	// Check if we're already in the cmd/grpcserver directory
	if _, err := os.Stat(".env"); err == nil {
		configDir = "."
	}

	var envFile string
	switch env {
	case "test":
		envFile = filepath.Join(configDir, ".env.test")
	case "production":
		envFile = filepath.Join(configDir, ".env.production")
	default: // development
		envFile = filepath.Join(configDir, ".env")
	}

	// Check if file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		fmt.Printf("⚠️  Warning: Environment file %s not found, using environment variables only\n", envFile)
		return nil
	}

	// Use viper to load the .env file
	viper.SetConfigFile(envFile)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file %s: %w", envFile, err)
	}

	// Set environment variables from viper
	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error setting env var %s: %w", key, err)
		}
	}

	fmt.Printf("✅ Loaded configuration from: %s (Environment: %s)\n", envFile, env)
	return nil
}
