package bootstrap

import (
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	platformES "github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch"
)

func InitElasticsearch(cfg *config.Config) (*platformES.Factory, error) {
	log.Printf("Connecting to Elasticsearch (%v)...", cfg.Elasticsearch.Addresses)

	factory, err := platformES.NewFactory(platformES.Config{
		Addresses:    cfg.Elasticsearch.Addresses,
		MaxRetries:   cfg.Elasticsearch.MaxRetries,
		RetryBackoff: cfg.Elasticsearch.RetryBackoff,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize elasticsearch: %w", err)
	}

	log.Println("Elasticsearch connected successfully")

	return factory, nil
}
