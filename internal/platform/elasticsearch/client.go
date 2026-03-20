package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type Config struct {
	Addresses    []string
	CloudID      string
	Username     string
	Password     string
	MaxRetries   int
	RetryBackoff time.Duration
}

func NewClient(cfg Config) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		MaxRetries:    cfg.MaxRetries,
		RetryBackoff:  func(i int) time.Duration { return cfg.RetryBackoff * time.Duration(i) },
		EnableMetrics: true,
	}

	if len(cfg.Addresses) > 0 {
		esCfg.Addresses = cfg.Addresses
	} else if cfg.CloudID != "" {
		esCfg.CloudID = cfg.CloudID
	}

	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch ping returned error: %s", res.Status())
	}

	return client, nil
}
