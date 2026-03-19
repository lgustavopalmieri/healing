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
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
	t.Setenv("ELASTICSEARCH_ADDRESSES", "http://es:9200")
	t.Setenv("LICENSE_VALIDATION_BASE_URL", "http://license-service:8080")
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
				assert.Equal(t, "kafka:9092", cfg.Kafka.BootstrapServers)
				assert.Equal(t, []string{"http://es:9200"}, cfg.Elasticsearch.Addresses)
				assert.Equal(t, "http://license-service:8080", cfg.External.LicenseBaseURL)
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
				assert.Equal(t, "earliest", cfg.Kafka.AutoOffsetReset)
				assert.Equal(t, 3, cfg.Elasticsearch.MaxRetries)
				assert.Equal(t, 1*time.Second, cfg.Elasticsearch.RetryBackoff)
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
				t.Setenv("KAFKA_AUTO_OFFSET_RESET", "latest")
				t.Setenv("ELASTICSEARCH_MAX_RETRIES", "5")
				t.Setenv("ELASTICSEARCH_RETRY_BACKOFF", "2s")
			},
			expectError: false,
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 50052, cfg.Server.GRPCPort)
				assert.Equal(t, 8081, cfg.Server.HTTPPort)
				assert.Equal(t, 60*time.Second, cfg.Server.ShutdownTimeout)
				assert.Equal(t, 2000, cfg.Server.MaxConnections)
				assert.Equal(t, "latest", cfg.Kafka.AutoOffsetReset)
				assert.Equal(t, 5, cfg.Elasticsearch.MaxRetries)
				assert.Equal(t, 2*time.Second, cfg.Elasticsearch.RetryBackoff)
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
				assert.Equal(t, 25, cfg.Database.MaxOpenConns)
				assert.Equal(t, 5, cfg.Database.MaxIdleConns)
				assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxLifetime)
				assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxIdleTime)
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
