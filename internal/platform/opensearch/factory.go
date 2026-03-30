package opensearch

import (
	"context"
	"fmt"
	"time"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch/indexes"
)

type Factory struct {
	Client      *opensearchapi.Client
	IndexPrefix string
}

func NewFactory(cfg Config) (*Factory, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := indexes.NewIndexRegistry(cfg.IndexPrefix)
	registry.Register(indexes.SpecialistsIndex, indexes.CreateSpecialistsIndex)

	if err := registry.CreateAll(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create opensearch indices: %w", err)
	}

	return &Factory{
		Client:      client,
		IndexPrefix: cfg.IndexPrefix,
	}, nil
}

func (f *Factory) IndexName(base string) string {
	if f.IndexPrefix == "" {
		return base
	}
	return f.IndexPrefix + "-" + base
}
