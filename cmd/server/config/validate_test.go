package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validConfigFactory(overrides ...func(*Config)) *Config {
	cfg := &Config{
		Server: ServerConfig{
			GRPCPort:          50051,
			HTTPPort:          8080,
			ShutdownTimeout:   30 * time.Second,
			MaxConnections:    1000,
			ConnectionTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "postgres.healing.svc.cluster.local",
			Port:            5432,
			User:            "app_user",
			Password:        "secret",
			Database:        "healing_specialist_db",
			SSLMode:         "require",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		SQS: SQSConfig{
			Region:      "us-east-1",
			QueuePrefix: "specialist",
		},
		OpenSearch: OpenSearchConfig{
			Addresses: []string{"http://opensearch.healing.svc.cluster.local:9200"},
		},
		External: ExternalConfig{
			LicenseBaseURL: "http://license-service.healing.svc.cluster.local:8080",
		},
	}

	for _, override := range overrides {
		override(cfg)
	}

	return cfg
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		override    func(*Config)
		expectError bool
		expectedMsg string
	}{
		{
			name:        "success - valid config passes validation",
			override:    func(c *Config) {},
			expectError: false,
		},
		{
			name:        "failure - returns error when GRPC port is zero",
			override:    func(c *Config) { c.Server.GRPCPort = 0 },
			expectError: true,
			expectedMsg: "invalid GRPC port: 0",
		},
		{
			name:        "failure - returns error when GRPC port exceeds 65535",
			override:    func(c *Config) { c.Server.GRPCPort = 70000 },
			expectError: true,
			expectedMsg: "invalid GRPC port: 70000",
		},
		{
			name:        "failure - returns error when HTTP port is zero",
			override:    func(c *Config) { c.Server.HTTPPort = 0 },
			expectError: true,
			expectedMsg: "invalid HTTP port: 0",
		},
		{
			name:        "failure - returns error when POSTGRES_HOST is empty",
			override:    func(c *Config) { c.Database.Host = "" },
			expectError: true,
			expectedMsg: "POSTGRES_HOST is required",
		},
		{
			name:        "failure - returns error when POSTGRES_USER is empty",
			override:    func(c *Config) { c.Database.User = "" },
			expectError: true,
			expectedMsg: "POSTGRES_USER is required",
		},
		{
			name:        "failure - returns error when POSTGRES_PASSWORD is empty",
			override:    func(c *Config) { c.Database.Password = "" },
			expectError: true,
			expectedMsg: "POSTGRES_PASSWORD is required",
		},
		{
			name:        "failure - returns error when POSTGRES_DB is empty",
			override:    func(c *Config) { c.Database.Database = "" },
			expectError: true,
			expectedMsg: "POSTGRES_DB is required",
		},
		{
			name:        "failure - returns error when SQS_REGION is empty",
			override:    func(c *Config) { c.SQS.Region = "" },
			expectError: true,
			expectedMsg: "SQS_REGION is required",
		},
		{
			name: "failure - returns error when OPENSEARCH_ADDRESSES is empty",
			override: func(c *Config) {
				c.OpenSearch.Addresses = nil
			},
			expectError: true,
			expectedMsg: "OPENSEARCH_ADDRESSES is required",
		},
		{
			name:        "success - valid config with empty LICENSE_VALIDATION_BASE_URL",
			override:    func(c *Config) { c.External.LicenseBaseURL = "" },
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfigFactory(tt.override)

			err := cfg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
