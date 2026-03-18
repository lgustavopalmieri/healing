package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

type IndexNames struct {
	Specialists string
}

type Factory struct {
	Client  *elasticsearch.Client
	Indexes IndexNames
}

func NewFactory(cfg Config) (*Factory, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	idx := IndexNames{
		Specialists: cfg.IndexSpecialists,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := indexes.NewIndexRegistry()
	registry.Register(idx.Specialists, indexes.CreateSpecialistsIndex)

	if err := registry.CreateAll(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch indices: %w", err)
	}

	return &Factory{
		Client:  client,
		Indexes: idx,
	}, nil
}
