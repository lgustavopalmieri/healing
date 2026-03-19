package config

import (
	"time"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Kafka         KafkaConfig
	Elasticsearch ElasticsearchConfig
	External      ExternalConfig
}

type ServerConfig struct {
	GRPCPort          int
	HTTPPort          int
	ShutdownTimeout   time.Duration
	MaxConnections    int
	ConnectionTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type KafkaConfig struct {
	BootstrapServers string
	AutoOffsetReset  string
}

type ElasticsearchConfig struct {
	Addresses    []string
	MaxRetries   int
	RetryBackoff time.Duration
}

type ExternalConfig struct {
	LicenseBaseURL string
}
