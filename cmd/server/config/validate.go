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

	if c.AuthDB.Host == "" {
		return fmt.Errorf("AUTH_POSTGRES_HOST is required")
	}

	if c.AuthDB.User == "" {
		return fmt.Errorf("AUTH_POSTGRES_USER is required")
	}

	if c.AuthDB.Password == "" {
		return fmt.Errorf("AUTH_POSTGRES_PASSWORD is required")
	}

	if c.AuthDB.Database == "" {
		return fmt.Errorf("AUTH_POSTGRES_DB is required")
	}

	if c.SQS.Region == "" {
		return fmt.Errorf("SQS_REGION is required")
	}

	if c.SNS.Region == "" {
		return fmt.Errorf("SNS_REGION is required")
	}

	if c.SNS.TopicPrefix == "" {
		return fmt.Errorf("SNS_TOPIC_PREFIX is required")
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

	if c.Auth.PrivateKeyPath == "" {
		return fmt.Errorf("AUTH_PRIVATE_KEY_PATH is required")
	}

	if c.Auth.PublicKeyPath == "" {
		return fmt.Errorf("AUTH_PUBLIC_KEY_PATH is required")
	}

	if c.Auth.CurrentKeyID == "" {
		return fmt.Errorf("AUTH_CURRENT_KEY_ID is required")
	}

	if c.Auth.Issuer == "" {
		return fmt.Errorf("AUTH_ISSUER is required")
	}

	if c.Auth.Audience == "" {
		return fmt.Errorf("AUTH_AUDIENCE is required")
	}

	if c.Auth.AccessTokenTTL <= 0 {
		return fmt.Errorf("AUTH_ACCESS_TOKEN_TTL must be > 0")
	}

	if c.Auth.RefreshTokenTTL <= 0 {
		return fmt.Errorf("AUTH_REFRESH_TOKEN_TTL must be > 0")
	}

	if c.Auth.SetPasswordTTL <= 0 {
		return fmt.Errorf("AUTH_SET_PASSWORD_TTL must be > 0")
	}

	if c.Auth.ResetPasswordTTL <= 0 {
		return fmt.Errorf("AUTH_RESET_PASSWORD_TTL must be > 0")
	}

	if c.Auth.BcryptCost < 10 || c.Auth.BcryptCost > 14 {
		return fmt.Errorf("AUTH_BCRYPT_COST must be between 10 and 14")
	}

	if c.Auth.PasswordMinLength < 8 {
		return fmt.Errorf("AUTH_PASSWORD_MIN_LENGTH must be >= 8")
	}

	return nil
}
