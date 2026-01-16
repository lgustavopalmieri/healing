package config

import (
	"fmt"
	"time"
)

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
