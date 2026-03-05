package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type Repository struct {
	client    *elasticsearch.Client
	indexName string
	logger    observability.Logger
}

func NewRepository(client *elasticsearch.Client, indexName string, logger observability.Logger) *Repository {
	return &Repository{
		client:    client,
		indexName: indexName,
		logger:    logger,
	}
}
