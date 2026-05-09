package config

import "time"

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	AuthDB     DatabaseConfig
	Redis      RedisConfig
	Auth       AuthConfig
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

type AuthConfig struct {
	PrivateKeyPath    string
	PublicKeyPath     string
	CurrentKeyID      string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	SetPasswordTTL    time.Duration
	ResetPasswordTTL  time.Duration
	Issuer            string
	Audience          string
	BcryptCost        int
	PasswordMinLength int
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}
