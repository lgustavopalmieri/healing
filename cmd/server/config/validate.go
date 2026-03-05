package config

import "fmt"

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

	if len(c.Elasticsearch.Addresses) == 0 {
		return fmt.Errorf("elasticsearch addresses is required")
	}

	if c.Elasticsearch.IndexSpecialists == "" {
		return fmt.Errorf("elasticsearch index for specialists is required")
	}

	return nil
}
