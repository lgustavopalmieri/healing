package config

import "fmt"

func (c *Config) Validate() error {
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC port: %d", c.Server.GRPCPort)
	}

	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Server.HTTPPort)
	}

	if c.Server.MetricsPort <= 0 || c.Server.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port: %d", c.Server.MetricsPort)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("POSTGRES_HOST is required")
	}

	if c.Database.User == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("POSTGRES_DB is required")
	}

	if c.Kafka.BootstrapServers == "" {
		return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required")
	}

	if c.Observability.ServiceName == "" {
		return fmt.Errorf("OTEL_SERVICE_NAME is required")
	}

	if c.Observability.ServiceVersion == "" {
		return fmt.Errorf("OTEL_SERVICE_VERSION is required")
	}

	if c.Observability.OTLPEndpoint == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_GRPC_ENDPOINT is required")
	}

	if len(c.Elasticsearch.Addresses) == 0 {
		return fmt.Errorf("ELASTICSEARCH_ADDRESSES is required")
	}

	if c.Elasticsearch.IndexSpecialists == "" {
		return fmt.Errorf("ELASTICSEARCH_INDEX_SPECIALISTS is required")
	}

	if c.External.LicenseBaseURL == "" {
		return fmt.Errorf("LICENSE_VALIDATION_BASE_URL is required")
	}

	return nil
}
