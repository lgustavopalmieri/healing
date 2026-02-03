package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	platformES "github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

func InitElasticsearch(cfg *config.Config) (*elasticsearch.Client, error) {
	log.Printf("🔍 Connecting to Elasticsearch (%v)...", cfg.Elasticsearch.Addresses)

	client, err := platformES.NewClient(platformES.Config{
		Addresses:        cfg.Elasticsearch.Addresses,
		MaxRetries:       cfg.Elasticsearch.MaxRetries,
		RetryBackoff:     cfg.Elasticsearch.RetryBackoff,
		IndexSpecialists: cfg.Elasticsearch.IndexSpecialists,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize elasticsearch: %w", err)
	}

	log.Println("🔄 Creating Elasticsearch indices...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := indexes.GetDefaultRegistry(cfg.Elasticsearch.IndexSpecialists)
	if err := registry.CreateAll(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create indices: %w", err)
	}

	log.Println("✅ Elasticsearch connected successfully")

	return client, nil
}
