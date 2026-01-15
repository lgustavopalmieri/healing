package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Kafka         KafkaConfig
	Observability ObservabilityConfig
}

type ServerConfig struct {
	GRPCPort          int
	ShutdownTimeout   time.Duration
	MaxConnections    int
	ConnectionTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type KafkaConfig struct {
	BootstrapServers string
	AutoOffsetReset  string
}

type ObservabilityConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	OTLPProtocol   string
}

func Load() (*Config, error) {
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
			Environment:    getEnv("OTEL_ENVIRONMENT", "development"),
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_GRPC_ENDPOINT", ""),
			OTLPProtocol:   getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC port: %d", c.Server.GRPCPort)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Kafka.BootstrapServers == "" {
		return fmt.Errorf("kafka bootstrap servers is required")
	}

	if c.Observability.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if c.Observability.OTLPEndpoint == "" {
		return fmt.Errorf("OTLP endpoint is required")
	}

	return nil
}

// Helper functions to read environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
