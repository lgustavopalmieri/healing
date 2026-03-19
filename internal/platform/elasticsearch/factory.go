package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

type Factory struct {
	Client *elasticsearch.Client
}

func NewFactory(cfg Config) (*Factory, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := indexes.NewIndexRegistry()
	registry.Register(indexes.SpecialistsIndex, indexes.CreateSpecialistsIndex)

	if err := registry.CreateAll(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch indices: %w", err)
	}

	return &Factory{
		Client: client,
	}, nil
}
