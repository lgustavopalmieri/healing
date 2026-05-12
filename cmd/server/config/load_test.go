package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setAllRequiredEnvVars(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres.healing.svc.cluster.local")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "app_user")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "healing_specialist_db")
	t.Setenv("POSTGRES_SSLMODE", "require")
	t.Setenv("AUTH_POSTGRES_HOST", "postgres-auth.healing.svc.cluster.local")
	t.Setenv("AUTH_POSTGRES_PORT", "5432")
	t.Setenv("AUTH_POSTGRES_USER", "healing_auth")
	t.Setenv("AUTH_POSTGRES_PASSWORD", "healing_auth")
	t.Setenv("AUTH_POSTGRES_DB", "healing_auth")
	t.Setenv("AUTH_POSTGRES_SSLMODE", "disable")
	t.Setenv("SQS_REGION", "us-east-1")
	t.Setenv("OPENSEARCH_ADDRESSES", "http://opensearch:9200")
	t.Setenv("LICENSE_VALIDATION_BASE_URL", "http://license-service:8080")
	t.Setenv("REDIS_HOST", "redis.healing.svc.cluster.local")
	t.Setenv("AUTH_PRIVATE_KEY_PATH", "/etc/healing/keys/auth-private.pem")
	t.Setenv("AUTH_PUBLIC_KEY_PATH", "/etc/healing/keys/auth-public.pem")
	t.Setenv("AUTH_CURRENT_KEY_ID", "healing-2026-05")
	t.Setenv("ENV_DIR", "/nonexistent")
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func(t *testing.T)
		expectError    bool
		expectedErrMsg string
		validateResult func(t *testing.T, cfg *Config)
	}{
		{
			name: "success - loads config from env vars with all required fields",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres.healing.svc.cluster.local", cfg.Database.Host)
				assert.Equal(t, 5432, cfg.Database.Port)
				assert.Equal(t, "app_user", cfg.Database.User)
				assert.Equal(t, "secret", cfg.Database.Password)
				assert.Equal(t, "healing_specialist_db", cfg.Database.Database)
				assert.Equal(t, "require", cfg.Database.SSLMode)
				assert.Equal(t, "postgres-auth.healing.svc.cluster.local", cfg.AuthDB.Host)
				assert.Equal(t, 5432, cfg.AuthDB.Port)
				assert.Equal(t, "healing_auth", cfg.AuthDB.User)
				assert.Equal(t, "healing_auth", cfg.AuthDB.Password)
				assert.Equal(t, "healing_auth", cfg.AuthDB.Database)
				assert.Equal(t, "disable", cfg.AuthDB.SSLMode)
				assert.Equal(t, "us-east-1", cfg.SQS.Region)
				assert.Equal(t, "specialist", cfg.SQS.QueuePrefix)
				assert.Equal(t, []string{"http://opensearch:9200"}, cfg.OpenSearch.Addresses)
				assert.Equal(t, "http://license-service:8080", cfg.External.LicenseBaseURL)
				assert.Equal(t, "redis.healing.svc.cluster.local", cfg.Redis.Host)
				assert.Equal(t, 6379, cfg.Redis.Port)
				assert.Equal(t, "", cfg.Redis.Password)
				assert.Equal(t, 0, cfg.Redis.DB)
				assert.Equal(t, 10, cfg.Redis.PoolSize)
				assert.Equal(t, 2, cfg.Redis.MinIdleConns)
			},
		},
		{
			name: "success - uses default values for optional fields when not set",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 50051, cfg.Server.GRPCPort)
				assert.Equal(t, 8080, cfg.Server.HTTPPort)
				assert.Equal(t, 30*time.Second, cfg.Server.ShutdownTimeout)
				assert.Equal(t, 1000, cfg.Server.MaxConnections)
				assert.Equal(t, 10*time.Second, cfg.Server.ConnectionTimeout)
				assert.Equal(t, "", cfg.SQS.Endpoint)
				assert.Equal(t, "", cfg.OpenSearch.Region)
				assert.Equal(t, "", cfg.OpenSearch.IndexPrefix)
			},
		},
		{
			name: "success - env vars take precedence over defaults",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
				t.Setenv("APP_ENV", "staging")
				t.Setenv("SERVER_GRPC_PORT", "50052")
				t.Setenv("SERVER_HTTP_PORT", "8081")
				t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "60s")
				t.Setenv("SERVER_MAX_CONNECTIONS", "2000")
				t.Setenv("SQS_QUEUE_PREFIX", "specialist-staging")
				t.Setenv("SQS_ENDPOINT", "http://localstack:4566")
				t.Setenv("OPENSEARCH_REGION", "us-east-1")
				t.Setenv("OPENSEARCH_INDEX_PREFIX", "healing")
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 50052, cfg.Server.GRPCPort)
				assert.Equal(t, 8081, cfg.Server.HTTPPort)
				assert.Equal(t, 60*time.Second, cfg.Server.ShutdownTimeout)
				assert.Equal(t, 2000, cfg.Server.MaxConnections)
				assert.Equal(t, "specialist-staging", cfg.SQS.QueuePrefix)
				assert.Equal(t, "http://localstack:4566", cfg.SQS.Endpoint)
				assert.Equal(t, "us-east-1", cfg.OpenSearch.Region)
				assert.Equal(t, "healing", cfg.OpenSearch.IndexPrefix)
			},
		},
		{
			name: "failure - returns error when required env vars are missing",
			setupEnv: func(t *testing.T) {
				t.Setenv("ENV_DIR", "/nonexistent")
			},
			expectError:    true,
			expectedErrMsg: "invalid configuration",
		},
		{
			name: "success - database pool config reads from env vars",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
				t.Setenv("POSTGRES_MAX_OPEN_CONNS", "50")
				t.Setenv("POSTGRES_MAX_IDLE_CONNS", "10")
				t.Setenv("POSTGRES_CONN_MAX_LIFETIME", "10m")
				t.Setenv("POSTGRES_CONN_MAX_IDLE_TIME", "20m")
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 50, cfg.Database.MaxOpenConns)
				assert.Equal(t, 10, cfg.Database.MaxIdleConns)
				assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxLifetime)
				assert.Equal(t, 20*time.Minute, cfg.Database.ConnMaxIdleTime)
			},
		},
		{
			name: "success - database pool uses defaults when env vars not set",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10, cfg.Database.MaxOpenConns)
				assert.Equal(t, 10, cfg.Database.MaxIdleConns)
				assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxLifetime)
				assert.Equal(t, 2*time.Minute, cfg.Database.ConnMaxIdleTime)
			},
		},
		{
			name: "success - redis config reads from env vars",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
				t.Setenv("REDIS_HOST", "redis-override")
				t.Setenv("REDIS_PORT", "6380")
				t.Setenv("REDIS_PASSWORD", "s3cret")
				t.Setenv("REDIS_DB", "3")
				t.Setenv("REDIS_POOL_SIZE", "25")
				t.Setenv("REDIS_MIN_IDLE_CONNS", "5")
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "redis-override", cfg.Redis.Host)
				assert.Equal(t, 6380, cfg.Redis.Port)
				assert.Equal(t, "s3cret", cfg.Redis.Password)
				assert.Equal(t, 3, cfg.Redis.DB)
				assert.Equal(t, 25, cfg.Redis.PoolSize)
				assert.Equal(t, 5, cfg.Redis.MinIdleConns)
			},
		},
		{
			name: "success - auth database pool config reads from env vars",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
				t.Setenv("AUTH_POSTGRES_MAX_OPEN_CONNS", "20")
				t.Setenv("AUTH_POSTGRES_MAX_IDLE_CONNS", "8")
				t.Setenv("AUTH_POSTGRES_CONN_MAX_LIFETIME", "15m")
				t.Setenv("AUTH_POSTGRES_CONN_MAX_IDLE_TIME", "3m")
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 20, cfg.AuthDB.MaxOpenConns)
				assert.Equal(t, 8, cfg.AuthDB.MaxIdleConns)
				assert.Equal(t, 15*time.Minute, cfg.AuthDB.ConnMaxLifetime)
				assert.Equal(t, 3*time.Minute, cfg.AuthDB.ConnMaxIdleTime)
			},
		},
		{
			name: "success - auth database uses defaults when env vars not set",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10, cfg.AuthDB.MaxOpenConns)
				assert.Equal(t, 5, cfg.AuthDB.MaxIdleConns)
				assert.Equal(t, 5*time.Minute, cfg.AuthDB.ConnMaxLifetime)
				assert.Equal(t, 1*time.Minute, cfg.AuthDB.ConnMaxIdleTime)
			},
		},
		{
			name: "success - auth config reads from env vars with custom values",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
				t.Setenv("AUTH_PRIVATE_KEY_PATH", "/etc/healing/keys/priv.pem")
				t.Setenv("AUTH_PUBLIC_KEY_PATH", "/etc/healing/keys/pub.pem")
				t.Setenv("AUTH_CURRENT_KEY_ID", "custom-kid")
				t.Setenv("AUTH_ACCESS_TOKEN_TTL", "2h")
				t.Setenv("AUTH_REFRESH_TOKEN_TTL", "720h")
				t.Setenv("AUTH_SET_PASSWORD_TTL", "12h")
				t.Setenv("AUTH_RESET_PASSWORD_TTL", "30m")
				t.Setenv("AUTH_ISSUER", "custom-issuer")
				t.Setenv("AUTH_AUDIENCE", "custom-audience")
				t.Setenv("AUTH_BCRYPT_COST", "13")
				t.Setenv("AUTH_PASSWORD_MIN_LENGTH", "10")
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "/etc/healing/keys/priv.pem", cfg.Auth.PrivateKeyPath)
				assert.Equal(t, "/etc/healing/keys/pub.pem", cfg.Auth.PublicKeyPath)
				assert.Equal(t, "custom-kid", cfg.Auth.CurrentKeyID)
				assert.Equal(t, 2*time.Hour, cfg.Auth.AccessTokenTTL)
				assert.Equal(t, 720*time.Hour, cfg.Auth.RefreshTokenTTL)
				assert.Equal(t, 12*time.Hour, cfg.Auth.SetPasswordTTL)
				assert.Equal(t, 30*time.Minute, cfg.Auth.ResetPasswordTTL)
				assert.Equal(t, "custom-issuer", cfg.Auth.Issuer)
				assert.Equal(t, "custom-audience", cfg.Auth.Audience)
				assert.Equal(t, 13, cfg.Auth.BcryptCost)
				assert.Equal(t, 10, cfg.Auth.PasswordMinLength)
			},
		},
		{
			name: "success - auth config uses defaults when optional env vars not set",
			setupEnv: func(t *testing.T) {
				setAllRequiredEnvVars(t)
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1*time.Hour, cfg.Auth.AccessTokenTTL)
				assert.Equal(t, 168*time.Hour, cfg.Auth.RefreshTokenTTL)
				assert.Equal(t, 24*time.Hour, cfg.Auth.SetPasswordTTL)
				assert.Equal(t, 1*time.Hour, cfg.Auth.ResetPasswordTTL)
				assert.Equal(t, "healing-specialist", cfg.Auth.Issuer)
				assert.Equal(t, "healing-platform", cfg.Auth.Audience)
				assert.Equal(t, 12, cfg.Auth.BcryptCost)
				assert.Equal(t, 8, cfg.Auth.PasswordMinLength)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv(t)

			cfg, err := Load()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validateResult != nil {
				tt.validateResult(t, cfg)
			}
		})
	}
}
