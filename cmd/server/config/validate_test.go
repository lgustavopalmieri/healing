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
			MaxOpenConns:    10,
			MaxIdleConns:    10,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 2 * time.Minute,
		},
		AuthDB: DatabaseConfig{
			Host:            "postgres-auth.healing.svc.cluster.local",
			Port:            5432,
			User:            "healing_auth",
			Password:        "healing_auth",
			Database:        "healing_auth",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
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
		Redis: RedisConfig{
			Host:         "redis.healing.svc.cluster.local",
			Port:         6379,
			Password:     "",
			DB:           0,
			PoolSize:     10,
			MinIdleConns: 2,
		},
		Auth: AuthConfig{
			PrivateKeyPath:    "/etc/healing/keys/auth-private.pem",
			PublicKeyPath:     "/etc/healing/keys/auth-public.pem",
			CurrentKeyID:      "healing-2026-05",
			AccessTokenTTL:    1 * time.Hour,
			RefreshTokenTTL:   168 * time.Hour,
			SetPasswordTTL:    24 * time.Hour,
			ResetPasswordTTL:  1 * time.Hour,
			Issuer:            "healing-specialist",
			Audience:          "healing-platform",
			BcryptCost:        12,
			PasswordMinLength: 8,
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
			name:        "failure - returns error when AUTH_POSTGRES_HOST is empty",
			override:    func(c *Config) { c.AuthDB.Host = "" },
			expectError: true,
			expectedMsg: "AUTH_POSTGRES_HOST is required",
		},
		{
			name:        "failure - returns error when AUTH_POSTGRES_USER is empty",
			override:    func(c *Config) { c.AuthDB.User = "" },
			expectError: true,
			expectedMsg: "AUTH_POSTGRES_USER is required",
		},
		{
			name:        "failure - returns error when AUTH_POSTGRES_PASSWORD is empty",
			override:    func(c *Config) { c.AuthDB.Password = "" },
			expectError: true,
			expectedMsg: "AUTH_POSTGRES_PASSWORD is required",
		},
		{
			name:        "failure - returns error when AUTH_POSTGRES_DB is empty",
			override:    func(c *Config) { c.AuthDB.Database = "" },
			expectError: true,
			expectedMsg: "AUTH_POSTGRES_DB is required",
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
			name:        "failure - returns error when REDIS_HOST is empty",
			override:    func(c *Config) { c.Redis.Host = "" },
			expectError: true,
			expectedMsg: "REDIS_HOST is required",
		},
		{
			name:        "failure - returns error when REDIS_PORT is zero",
			override:    func(c *Config) { c.Redis.Port = 0 },
			expectError: true,
			expectedMsg: "invalid REDIS_PORT: 0",
		},
		{
			name:        "failure - returns error when REDIS_PORT exceeds 65535",
			override:    func(c *Config) { c.Redis.Port = 70000 },
			expectError: true,
			expectedMsg: "invalid REDIS_PORT: 70000",
		},
		{
			name:        "failure - returns error when REDIS_POOL_SIZE is less than 1",
			override:    func(c *Config) { c.Redis.PoolSize = 0 },
			expectError: true,
			expectedMsg: "REDIS_POOL_SIZE must be >= 1",
		},
		{
			name:        "failure - returns error when AUTH_PRIVATE_KEY_PATH is empty",
			override:    func(c *Config) { c.Auth.PrivateKeyPath = "" },
			expectError: true,
			expectedMsg: "AUTH_PRIVATE_KEY_PATH is required",
		},
		{
			name:        "failure - returns error when AUTH_PUBLIC_KEY_PATH is empty",
			override:    func(c *Config) { c.Auth.PublicKeyPath = "" },
			expectError: true,
			expectedMsg: "AUTH_PUBLIC_KEY_PATH is required",
		},
		{
			name:        "failure - returns error when AUTH_CURRENT_KEY_ID is empty",
			override:    func(c *Config) { c.Auth.CurrentKeyID = "" },
			expectError: true,
			expectedMsg: "AUTH_CURRENT_KEY_ID is required",
		},
		{
			name:        "failure - returns error when AUTH_ISSUER is empty",
			override:    func(c *Config) { c.Auth.Issuer = "" },
			expectError: true,
			expectedMsg: "AUTH_ISSUER is required",
		},
		{
			name:        "failure - returns error when AUTH_AUDIENCE is empty",
			override:    func(c *Config) { c.Auth.Audience = "" },
			expectError: true,
			expectedMsg: "AUTH_AUDIENCE is required",
		},
		{
			name:        "failure - returns error when AUTH_ACCESS_TOKEN_TTL is zero",
			override:    func(c *Config) { c.Auth.AccessTokenTTL = 0 },
			expectError: true,
			expectedMsg: "AUTH_ACCESS_TOKEN_TTL must be > 0",
		},
		{
			name:        "failure - returns error when AUTH_REFRESH_TOKEN_TTL is zero",
			override:    func(c *Config) { c.Auth.RefreshTokenTTL = 0 },
			expectError: true,
			expectedMsg: "AUTH_REFRESH_TOKEN_TTL must be > 0",
		},
		{
			name:        "failure - returns error when AUTH_SET_PASSWORD_TTL is zero",
			override:    func(c *Config) { c.Auth.SetPasswordTTL = 0 },
			expectError: true,
			expectedMsg: "AUTH_SET_PASSWORD_TTL must be > 0",
		},
		{
			name:        "failure - returns error when AUTH_RESET_PASSWORD_TTL is zero",
			override:    func(c *Config) { c.Auth.ResetPasswordTTL = 0 },
			expectError: true,
			expectedMsg: "AUTH_RESET_PASSWORD_TTL must be > 0",
		},
		{
			name:        "failure - returns error when AUTH_BCRYPT_COST below 10",
			override:    func(c *Config) { c.Auth.BcryptCost = 9 },
			expectError: true,
			expectedMsg: "AUTH_BCRYPT_COST must be between 10 and 14",
		},
		{
			name:        "failure - returns error when AUTH_BCRYPT_COST above 14",
			override:    func(c *Config) { c.Auth.BcryptCost = 15 },
			expectError: true,
			expectedMsg: "AUTH_BCRYPT_COST must be between 10 and 14",
		},
		{
			name:        "failure - returns error when AUTH_PASSWORD_MIN_LENGTH below 8",
			override:    func(c *Config) { c.Auth.PasswordMinLength = 7 },
			expectError: true,
			expectedMsg: "AUTH_PASSWORD_MIN_LENGTH must be >= 8",
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
