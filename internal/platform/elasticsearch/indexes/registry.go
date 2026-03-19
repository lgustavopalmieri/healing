package indexes

import (
	"context"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	SpecialistsIndex = "specialists"
)

type IndexCreator func(ctx context.Context, client *elasticsearch.Client, indexName string) error

type IndexRegistry struct {
	indexes map[string]IndexCreator
}

func NewIndexRegistry() *IndexRegistry {
	return &IndexRegistry{
		indexes: make(map[string]IndexCreator),
	}
}

func (r *IndexRegistry) Register(indexName string, creator IndexCreator) {
	r.indexes[indexName] = creator
}

func (r *IndexRegistry) CreateAll(ctx context.Context, client *elasticsearch.Client) error {
	for indexName, creator := range r.indexes {
		log.Printf("Creating index: %s", indexName)
		if err := creator(ctx, client, indexName); err != nil {
			return fmt.Errorf("failed to create index %s: %w", indexName, err)
		}
		log.Printf("Index created: %s", indexName)
	}
	return nil
}
