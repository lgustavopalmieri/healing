package config

import (
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
