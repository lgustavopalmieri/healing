package bootstrap

import (
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	platformOS "github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch"
)

func InitOpenSearch(cfg *config.Config) (*platformOS.Factory, error) {
	log.Printf("Connecting to OpenSearch (%v)...", cfg.OpenSearch.Addresses)

	factory, err := platformOS.NewFactory(platformOS.Config{
		Addresses:   cfg.OpenSearch.Addresses,
		Region:      cfg.OpenSearch.Region,
		IndexPrefix: cfg.OpenSearch.IndexPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize opensearch: %w", err)
	}

	log.Println("OpenSearch connected successfully")

	return factory, nil
}
