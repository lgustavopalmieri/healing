package config

import "fmt"

func (c *Config) Validate() error {
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC port: %d", c.Server.GRPCPort)
	}

	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Server.HTTPPort)
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

	if c.SQS.Region == "" {
		return fmt.Errorf("SQS_REGION is required")
	}

	if len(c.OpenSearch.Addresses) == 0 {
		return fmt.Errorf("OPENSEARCH_ADDRESSES is required")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}

	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return fmt.Errorf("invalid REDIS_PORT: %d", c.Redis.Port)
	}

	if c.Redis.PoolSize < 1 {
		return fmt.Errorf("REDIS_POOL_SIZE must be >= 1")
	}

	return nil
}
