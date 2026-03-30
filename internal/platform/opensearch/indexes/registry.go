package indexes

import (
	"context"
	"fmt"
	"log"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

const (
	SpecialistsIndex = "specialists"
)

type IndexCreator func(ctx context.Context, client *opensearchapi.Client, indexName string) error

type IndexRegistry struct {
	indexes map[string]IndexCreator
	prefix  string
}

func NewIndexRegistry(prefix string) *IndexRegistry {
	return &IndexRegistry{
		indexes: make(map[string]IndexCreator),
		prefix:  prefix,
	}
}

func (r *IndexRegistry) PrefixedName(name string) string {
	if r.prefix == "" {
		return name
	}
	return r.prefix + "-" + name
}

func (r *IndexRegistry) Register(indexName string, creator IndexCreator) {
	r.indexes[indexName] = creator
}

func (r *IndexRegistry) CreateAll(ctx context.Context, client *opensearchapi.Client) error {
	for indexName, creator := range r.indexes {
		fullName := r.PrefixedName(indexName)
		log.Printf("Creating index: %s", fullName)
		if err := creator(ctx, client, fullName); err != nil {
			return fmt.Errorf("failed to create index %s: %w", fullName, err)
		}
		log.Printf("Index created: %s", fullName)
	}
	return nil
}
