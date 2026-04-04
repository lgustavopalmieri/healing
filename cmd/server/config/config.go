package config

import "time"

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	SQS        SQSConfig
	OpenSearch OpenSearchConfig
	External   ExternalConfig
	Otel       OtelConfig
}

type OtelConfig struct {
	ExporterEndpoint   string
	ExporterProtocol   string
	ServiceName        string
	ResourceAttributes string
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

type SQSConfig struct {
	Region      string
	QueuePrefix string
	Endpoint    string
}

type OpenSearchConfig struct {
	Addresses   []string
	Region      string
	IndexPrefix string
}

type ExternalConfig struct {
	LicenseBaseURL string
}
