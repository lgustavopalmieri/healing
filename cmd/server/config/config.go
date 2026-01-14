package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
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
	Broker           string
}

type ObservabilityConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	OTLPProtocol   string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read .env file: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			GRPCPort:          viper.GetInt("SERVER_GRPC_PORT"),
			ShutdownTimeout:   viper.GetDuration("SERVER_SHUTDOWN_TIMEOUT"),
			MaxConnections:    viper.GetInt("SERVER_MAX_CONNECTIONS"),
			ConnectionTimeout: viper.GetDuration("SERVER_CONNECTION_TIMEOUT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetInt("POSTGRES_PORT"),
			User:     viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			Database: viper.GetString("POSTGRES_DB"),
		},
		Kafka: KafkaConfig{
			BootstrapServers: viper.GetString("KAFKA_BOOTSTRAP_SERVERS"),
			Broker:           viper.GetString("KAFKA_BROKER"),
		},
		Observability: ObservabilityConfig{
			ServiceName:    viper.GetString("OTEL_SERVICE_NAME"),
			ServiceVersion: viper.GetString("OTEL_SERVICE_VERSION"),
			Environment:    viper.GetString("OTEL_ENVIRONMENT"),
			OTLPEndpoint:   viper.GetString("OTEL_EXPORTER_OTLP_GRPC_ENDPOINT"),
			OTLPProtocol:   viper.GetString("OTEL_EXPORTER_OTLP_PROTOCOL"),
		},
	}

	if cfg.Kafka.Broker == "" {
		cfg.Kafka.Broker = cfg.Kafka.BootstrapServers
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
